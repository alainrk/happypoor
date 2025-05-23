package client

import (
	"cashout/internal/model"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

func (c *Client) FreeTextRouter(b *gotgbot.Bot, ctx *ext.Context) error {
	_, u := c.getUserFromContext(ctx)
	user, err := c.authAndGetUser(u)
	if err != nil {
		return err
	}

	msg := ctx.Message
	if message.Text(msg) && strings.ToLower(strings.Trim(msg.Text, " ")) == "cancel" {
		return c.Cancel(b, ctx)
	}

	if user.Session.State == model.StateInsertingIncome || user.Session.State == model.StateInsertingExpense {
		return c.addTransaction(b, ctx, user)
	}

	// During-insert edit transaction

	if user.Session.State == model.StateEditingTransactionDate {
		return c.editTransactionDate(b, ctx, user)
	}

	if user.Session.State == model.StateEditingTransactionCategory {
		return c.editTransactionCategory(b, ctx, user)
	}

	if user.Session.State == model.StateEditingTransactionAmount {
		return c.editTransactionAmount(b, ctx, user)
	}

	if user.Session.State == model.StateEditingTransactionDescription {
		return c.editTransactionDescription(b, ctx, user)
	}

	// End of during-insert edit transaction

	// Top-level edit transaction

	if user.Session.State == model.StateTopLevelEditingTransactionDate {
		return c.EditTransactionDateConfirm(b, ctx)
	}

	if user.Session.State == model.StateTopLevelEditingTransactionCategory {
		return c.EditTransactionCategoryConfirm(b, ctx)
	}

	if user.Session.State == model.StateTopLevelEditingTransactionAmount {
		return c.EditTransactionAmountConfirm(b, ctx)
	}

	if user.Session.State == model.StateTopLevelEditingTransactionDescription {
		return c.EditTransactionDescriptionConfirm(b, ctx)
	}

	// End of top-level edit transaction

	// Default behavior: start expense transaction flow for any unhandled text.
	// Heuristic: there must be at least a digit in text.
	if strings.ContainsAny(ctx.Message.Text, "0123456789") {
		user.Session.State = model.StateInsertingExpense
		err = c.Repositories.Users.Update(&user)
		if err != nil {
			return fmt.Errorf("failed to set user data: %w", err)
		}

		// Process as expense transaction
		return c.addTransaction(b, ctx, user)
	}

	c.CleanupKeyboard(b, ctx)
	c.SendHomeKeyboard(b, ctx, "Sorry I don't understand, what can I do for you?\n\n/edit - Edit a transaction\n/delete - Delete a transaction\n/list - List your transactions\n/week Week Recap\n/month Month Recap\n/year Year Recap\n/export - Export all transactions to CSV")

	return fmt.Errorf("invalid top-level state")
}
