package client

import (
	"encoding/json"
	"fmt"
	"happypoor/internal/model"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// Start introduces the bot.
func (c *Client) Start(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := c.authAndGetUser(ctx)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Welcome to HappyPoor, %s!", user.Name)
	ctx.EffectiveMessage.Reply(b, msg, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})

	ctx.EffectiveMessage.Reply(b, "Add a transaction", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{
					{
						Text: "Add Income",
					},
					{
						Text: "Add Expense",
					},
				},
			},
			IsPersistent:   true,
			ResizeKeyboard: true,
		},
	})

	return nil
}

func (c *Client) AddIncomeIntent(b *gotgbot.Bot, ctx *ext.Context) error {
	return c.addTransactionIntent(b, ctx, model.TypeIncome)
}

func (c *Client) AddExpenseIntent(b *gotgbot.Bot, ctx *ext.Context) error {
	return c.addTransactionIntent(b, ctx, model.TypeExpense)
}

func (c *Client) addTransactionIntent(b *gotgbot.Bot, ctx *ext.Context, transactionType model.TransactionType) error {
	user, err := c.authAndGetUser(ctx)
	if err != nil {
		return err
	}

	if user.Session.Iterations == 0 {
		return c.Start(b, ctx)
	}

	user.Session.LastCommand = model.CommandAddExpenseIntent
	if transactionType == model.TypeIncome {
		user.Session.LastCommand = model.CommandAddIncomeIntent
	}
	user.Session.State = model.StateWaiting
	user.Session.LastMessage = ctx.Message.Text

	err = c.Repositories.Users.Update(&user)
	if err != nil {
		return fmt.Errorf("failed to set user data: %w", err)
	}

	ctx.EffectiveMessage.Reply(b, "Sure, just tell me category, amount and decription.", &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{
					{
						Text: "Cancel",
					},
				},
			},
			IsPersistent:    false,
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		},
	})

	return nil
}

func (c *Client) AddTransaction(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := c.authAndGetUser(ctx)
	if err != nil {
		return err
	}

	if user.Session.Iterations == 0 {
		return c.Start(b, ctx)
	}

	var transactionType model.TransactionType

	switch user.Session.LastCommand {
	case model.CommandAddIncomeIntent:
		transactionType = model.TypeIncome
	case model.CommandAddExpenseIntent:
		transactionType = model.TypeExpense
	default:
		// answer the user that they should chose a valid command first and send the keyboard
		ctx.EffectiveMessage.Reply(b, "Add a transaction", &gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
				Keyboard: [][]gotgbot.KeyboardButton{
					{
						{
							Text: "Add Income",
						},
						{
							Text: "Add Expense",
						},
					},
				},
				IsPersistent:   true,
				ResizeKeyboard: true,
			},
		})
		return nil
	}

	transaction, err := c.LLM.ExtractTransaction(ctx.Message.Text, transactionType)
	if err != nil {
		msg := fmt.Sprintf("I'm sorry, I can't understand your transaction '%s', %s!", user.Session.LastMessage, user.Name)
		ctx.EffectiveMessage.Reply(b, msg, &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
		return err
	}

	// Store the transaction in the session
	user.Session.State = model.StateWaiting
	user.Session.LastCommand = model.CommandAddTransaction
	user.Session.LastMessage = ctx.Message.Text
	s, err := json.Marshal(transaction)
	user.Session.Body = string(s)
	if err != nil {
		return fmt.Errorf("failed to set user data: %w", err)
	}

	err = c.Repositories.Users.Update(&user)
	if err != nil {
		return fmt.Errorf("failed to set user data: %w", err)
	}

	msg := fmt.Sprintf("%s (€ %.2f), %s. Confirm?", transaction.Category, transaction.Amount, transaction.Description)
	ctx.EffectiveMessage.Reply(b, msg, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{
					{
						Text: "Confirm",
					},
					{
						Text: "Cancel",
					},
				},
			},
			IsPersistent:    false,
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})

	return nil
}

// Confirm confirms the previous action after the user been prompted.
func (c *Client) Confirm(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := c.authAndGetUser(ctx)
	if err != nil {
		return err
	}

	var transaction model.Transaction
	err = json.Unmarshal([]byte(user.Session.Body), &transaction)
	if err != nil {
		return fmt.Errorf("failed to extract transaction from the session: %w", err)
	}

	user.Session.State = model.StateNormal
	user.Session.LastCommand = model.CommandConfirm
	user.Session.LastMessage = ctx.Message.Text
	user.Session.Body = ""

	err = c.Repositories.Users.Update(&user)
	if err != nil {
		return fmt.Errorf("failed to set user data: %w", err)
	}

	ctx.EffectiveMessage.Reply(b, "Your transaction has been saved!", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{
					{
						Text: "Add Income",
					},
					{
						Text: "Add Expense",
					},
				},
			},
			IsPersistent:   true,
			ResizeKeyboard: true,
		},
	})

	return nil
}

// Cancel returns to normal state.
func (c *Client) Cancel(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := c.authAndGetUser(ctx)
	if err != nil {
		return err
	}

	user.Session.State = model.StateNormal
	user.Session.LastCommand = model.CommandCancel
	user.Session.LastMessage = ctx.Message.Text

	err = c.Repositories.Users.Update(&user)
	if err != nil {
		return fmt.Errorf("failed to set user data: %w", err)
	}

	ctx.EffectiveMessage.Reply(b, "Add a transaction", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{
					{
						Text: "Add Income",
					},
					{
						Text: "Add Expense",
					},
				},
			},
			IsPersistent:   true,
			ResizeKeyboard: true,
		},
	})

	return nil
}
