package services

import (
	"database/sql" // for sql.ErrNoRows
	"errors"
	"fmt"
	"net/http"
	"strings"
	"superapps/entities"
	helper "superapps/helpers"
	"time"
)

func ContractLetterPaymentUpload(r *http.Request, clpu *entities.ContractLetterPaymentUpload) (entities.ContractLetterPaymentUpload, error) {
	// Flexible SET builder (add more fields as needed)
	set := []string{}
	args := []any{}

	if clpu.Path != "" {
		set = append(set, "path = ?")
		args = append(args, clpu.Path)
	}

	// If nothing to update, just return (unchanged behavior)
	if len(set) == 0 {
		return entities.ContractLetterPaymentUpload{}, nil
	}

	// --- VALIDATE: does the row already exist? ---
	var one int
	scanErr := dbDefault.Raw(
		`SELECT 1 FROM contract_letter_payments WHERE payment_id = ? LIMIT 1`,
		clpu.ProjectPaymentId,
	).Row().Scan(&one)

	if scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
		helper.Logger("error", "In Server: "+scanErr.Error())
		return entities.ContractLetterPaymentUpload{}, scanErr
	}

	if scanErr == nil {
		// Exists -> UPDATE
		query := fmt.Sprintf(
			`UPDATE contract_letter_payments SET %s WHERE payment_id = ?`,
			strings.Join(set, ", "),
		)
		args = append(args, clpu.ProjectPaymentId)

		if err := dbDefault.Exec(query, args...).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.ContractLetterPaymentUpload{}, err
		}
	} else {
		// Not exists -> INSERT
		cols := []string{"payment_id"}
		vals := []string{"?"}
		insArgs := []any{clpu.ProjectPaymentId}

		if clpu.Path != "" {
			cols = append(cols, "path")
			vals = append(vals, "?")
			insArgs = append(insArgs, clpu.Path)
		}

		insQuery := fmt.Sprintf(
			`INSERT INTO contract_letter_payments (%s) VALUES (%s)`,
			strings.Join(cols, ", "),
			strings.Join(vals, ", "),
		)

		if err := dbDefault.Exec(insQuery, insArgs...).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.ContractLetterPaymentUpload{}, err
		}
	}

	email, title, role, _ := helper.GetEmailAndTitleAndRoleByContractLetterPayment(dbDefault, clpu.ProjectPaymentId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s mendapatkan pengiriman surat kontrak pembayaran pada project dengan judul [%s] pada waktu %s",
			ip,
			email,
			role,
			title,
			time.Now().Format(time.Now().Format("2006-01-02 15:04:05")),
		),
		role,
	)

	return entities.ContractLetterPaymentUpload{
		ProjectPaymentId: clpu.ProjectPaymentId,
		Path:             clpu.Path,
	}, nil
}
