package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ossyne/internal/models"
	"strconv"
	"github.com/spf13/cobra"
)

func NewPaymentCmd() *cobra.Command {
	walletCmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage bounties and withdrawals",
		Long:  `The wallet command provides tools for funding tasks, managing bounties, and viewing payment history.`,
	}

	fundCmd := &cobra.Command{
		Use:   "fund <task-id>",
		Short: "Fund a task with a bounty",
		Long:  `Place funds into escrow for a specific task to act as a bounty.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskIDStr := args[0]
			amountStr, _ := cmd.Flags().GetString("amount")
			currency, _ := cmd.Flags().GetString("currency")

			if amountStr == "" {
				fmt.Println("Error: --amount flag is required.")
				return
			}

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid task ID: %v\n", err)
				return
			}
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				fmt.Printf("Error: Invalid amount: %v\n", err)
				return
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"task_id":  uint(taskID),
				"amount":   amount,
				"currency": currency,
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPost, "/bounties/fund", payloadMap)
			if err != nil {
				fmt.Printf("Error funding task bounty: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error funding task bounty: %s\n", string(body))
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err == nil {
				if msg, ok := response["message"].(string); ok {
					fmt.Println(msg)
					return
				}
			}

			fmt.Printf("Successfully funded task %d with %.2f %s\n", taskID, amount, currency)
		},
	}
	fundCmd.Flags().StringP("amount", "a", "", "Amount of the bounty")
	fundCmd.Flags().StringP("currency", "c", "USD", "Currency of the bounty (default: USD)")
	fundCmd.MarkFlagRequired("amount")
	walletCmd.AddCommand(fundCmd)

	refundCmd := &cobra.Command{
		Use:   "refund <task-id>",
		Short: "Refund an escrowed bounty for a task",
		Long:  `Refunds the bounty for a task if it was not completed or accepted.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskIDStr := args[0]
			reason, _ := cmd.Flags().GetString("reason")

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid task ID: %v\n", err)
				return
			}

			apiClient := NewAPIClient()
			payloadMap := map[string]interface{}{
				"reason": reason,
			}

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodPut, fmt.Sprintf("/bounties/refund/%d", taskID), payloadMap)
			if err != nil {
				fmt.Printf("Error refunding task bounty: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error refunding task bounty: %s\n", string(body))
				return
			}

			fmt.Printf("Successfully refunded bounty for task %d\n", taskID)
		},
	}
	refundCmd.Flags().StringP("reason", "r", "No completion or agreement.", "Reason for refunding the bounty")
	walletCmd.AddCommand(refundCmd)

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "View your payment history",
		Long:  `Shows a list of all payment-related transactions for the authenticated user.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			apiClient := NewAPIClient()

			resp, err := apiClient.DoAuthenticatedRequest(http.MethodGet, "/users/me/payments", nil)
			if err != nil {
				fmt.Printf("Error fetching payment history: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error fetching payment history: %s\n", string(body))
				return
			}

			var payments []models.Payment
			if err := json.Unmarshal(body, &payments); err != nil {
				fmt.Printf("Error parsing payment history: %v\n", err)
				return
			}

			if len(payments) == 0 {
				fmt.Println("No payment history found.")
				return
			}

			fmt.Println("--- Payment History ---")
			for _, p := range payments {
				fmt.Printf("ID: %d, Amount: %.2f %s, Status: %s, Type: %s, Date: %s\n",
					p.ID, p.Amount, p.Currency, p.Status, p.Type, p.PaymentDate.Format("2006-01-02"))
			}
		},
	}
	walletCmd.AddCommand(historyCmd)

	return walletCmd
}