package cli

import (
	"bytes"
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
			funderUserIDStr, _ := cmd.Flags().GetString("funder-user-id")
			amountStr, _ := cmd.Flags().GetString("amount")
			currency, _ := cmd.Flags().GetString("currency")

			if funderUserIDStr == "" || amountStr == "" {
				fmt.Println("Error: --funder-user-id and --amount flags are required.")
				return
			}

			taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
			if err != nil { fmt.Printf("Error: Invalid task ID: %v\n", err); return }
			funderUserID, err := strconv.ParseUint(funderUserIDStr, 10, 64)
			if err != nil { fmt.Printf("Error: Invalid funder-user-id: %v\n", err); return }
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil { fmt.Printf("Error: Invalid amount: %v\n", err); return }

			fundTaskBounty(uint(taskID), uint(funderUserID), amount, currency)
		},
	}
	fundCmd.Flags().StringP("funder-user-id", "u", "", "ID of the user funding the bounty (e.g., project owner or sponsor)")
	fundCmd.Flags().StringP("amount", "a", "", "Amount of the bounty")
	fundCmd.Flags().StringP("currency", "c", "USD", "Currency of the bounty (default: USD)")
	fundCmd.MarkFlagRequired("funder-user-id")
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
			if err != nil { fmt.Printf("Error: Invalid task ID: %v\n", err); return }

			refundTaskBounty(uint(taskID), reason)
		},
	}
	refundCmd.Flags().StringP("reason", "r", "No completion or agreement.", "Reason for refunding the bounty")
	walletCmd.AddCommand(refundCmd)

	historyCmd := &cobra.Command{
		Use:   "history <user-id>",
		Short: "View payment history for a user",
		Long:  `Shows a list of all payment-related transactions for a specific user.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			userIDStr := args[0]
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil { fmt.Printf("Error: Invalid user ID: %v\n", err); return }
			viewPaymentHistory(uint(userID))
		},
	}
	walletCmd.AddCommand(historyCmd)

	return walletCmd
}

func fundTaskBounty(taskID, funderUserID uint, amount float64, currency string) {
	const serverURL = "http://localhost:8080/bounties/fund"

	payload, err := json.Marshal(map[string]interface{}{
		"task_id":        taskID,
		"funder_user_id": funderUserID,
		"amount":         amount,
		"currency":       currency,
	})
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		return
	}
	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", serverURL)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Task bounty funded and escrowed successfully!")
	} else {
		fmt.Printf("Error: Failed to fund task bounty (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func refundTaskBounty(taskID uint, reason string) {
	url := fmt.Sprintf("http://localhost:8080/bounties/refund/%d", taskID)

	payload, err := json.Marshal(map[string]string{
		"reason": reason,
	})
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		return
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", url)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Task bounty refunded successfully!")
	} else {
		fmt.Printf("Error: Failed to refund task bounty (Status: %s)\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func viewPaymentHistory(userID uint) {
	url := fmt.Sprintf("http://localhost:8080/users/%d/payments", userID)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: Could not connect to the OSM server at %s. Is it running?\n", url)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Failed to retrieve payment history (Status: %s)\n", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading server response: %v\n", err)
		return
	}

	var payments []models.Payment
	if err := json.Unmarshal(body, &payments); err != nil {
		fmt.Printf("Error parsing server response: %v\n", err)
		return
	}

	if len(payments) == 0 {
		fmt.Printf("No payment history found for user ID %d.\n", userID)
		return
	}

	fmt.Printf("--- Payment History for User ID %d ---\n", userID)
	for _, p := range payments {
		contribInfo := ""
		if p.ContributionID != nil {
			contribInfo = fmt.Sprintf(" (Contrib ID: %d)", *p.ContributionID)
		}
		fmt.Printf("ID: %d, Type: %s, Amount: %.2f %s, Status: %s, Date: %s%s\n",
			p.ID, p.Type, p.Amount, p.Currency, p.Status, p.PaymentDate.Format("2006-01-02"), contribInfo)
	}
}