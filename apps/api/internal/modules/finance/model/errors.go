package model

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

var (
	// Context
	ErrContextNotActive = errors.New("FINANCE_CONTEXT_NOT_ACTIVE")

	// Account
	ErrAccountTypeWithPostings      = errors.New("ACCOUNT_TYPE_WITH_POSTINGS")
	ErrAccountPostableWithPostings  = errors.New("ACCOUNT_POSTABLE_WITH_POSTINGS")
	ErrAccountDescendantHasPostings = errors.New("ACCOUNT_DESCENDANT_HAS_POSTINGS")
	ErrAccountPostableWithChildren  = errors.New("ACCOUNT_POSTABLE_WITH_CHILDREN")
	ErrAccountDeleteWithChildren    = errors.New("ACCOUNT_DELETE_WITH_CHILDREN")
	ErrAccountDeleteWithPostings    = errors.New("ACCOUNT_DELETE_WITH_POSTINGS")
	ErrAccountCodeExists            = errors.New("ACCOUNT_CODE_EXISTS")
	ErrParentAccountNotExist        = errors.New("PARENT_ACCOUNT_NOT_EXIST")
	ErrParentAccountPostable        = errors.New("PARENT_ACCOUNT_POSTABLE")
	ErrAccountTypeMismatch          = errors.New("ACCOUNT_TYPE_MISMATCH")
	ErrAccountSelfParent            = errors.New("ACCOUNT_SELF_PARENT")
	ErrAccountCycleDetected         = errors.New("ACCOUNT_CYCLE_DETECTED")

	// Activate
	ErrParishNotFound    = errors.New("PARISH_NOT_FOUND")
	ErrWorkspaceNotFound = errors.New("WORKSPACE_NOT_FOUND")

	// Report
	ErrAccountIdRequired           = errors.New("ACCOUNT_ID_REQUIRED")
	ErrAccountNotFoundInParish     = errors.New("ACCOUNT_NOT_FOUND_IN_PARISH")
	ErrDatesRequired               = errors.New("DATES_REQUIRED")
	ErrInvalidFromDate             = errors.New("INVALID_FROM_DATE")
	ErrInvalidToDate               = errors.New("INVALID_TO_DATE")
	ErrFromDateAfterToDate         = errors.New("FROM_DATE_AFTER_TO_DATE")
	ErrInvalidCursor               = errors.New("INVALID_CURSOR")
	ErrAsOfDateRequired            = errors.New("AS_OF_DATE_REQUIRED")
	ErrInvalidAsOfDate             = errors.New("INVALID_AS_OF_DATE")
	ErrDateRangeExceedsLimit       = errors.New("DATE_RANGE_EXCEEDS_LIMIT")
	ErrParentAccountNotFound       = errors.New("PARENT_ACCOUNT_NOT_FOUND")
	ErrParentMustBeHeader          = errors.New("PARENT_MUST_BE_HEADER")
	ErrInvalidAccountIDParam       = errors.New("INVALID_ACCOUNT_ID_PARAM")
	ErrInvalidParentAccountIDParam = errors.New("INVALID_PARENT_ACCOUNT_ID_PARAM")

	// Transaction
	ErrInvalidDateFormat         = errors.New("INVALID_DATE_FORMAT")
	ErrInvalidTxNumberFormat     = errors.New("INVALID_TX_NUMBER_FORMAT")
	ErrInvalidPostingSide        = errors.New("INVALID_POSTING_SIDE")
	ErrTxMissingPostings         = errors.New("TX_MISSING_POSTINGS")
	ErrTxUnbalanced              = errors.New("TX_UNBALANCED")
	ErrAccountsNotFound          = errors.New("ACCOUNTS_NOT_FOUND")
	ErrAccountNotActive          = errors.New("ACCOUNT_NOT_ACTIVE")
	ErrAccountIsHeader           = errors.New("ACCOUNT_IS_HEADER")
	ErrTransactionHasNoWorkspace = errors.New("TRANSACTION_HAS_NO_WORKSPACE")
	ErrTxNumberExists            = errors.New("TX_NUMBER_EXISTS")
	ErrTxNumberAllocationFailed  = errors.New("TX_NUMBER_ALLOCATION_FAILED")
	ErrTxNotPosted               = errors.New("TX_NOT_POSTED")
)

func init() {
	// Context
	core.RegisterError(ErrContextNotActive, fiber.StatusBadRequest, "finance context is not active, please activate first")

	// Account
	core.RegisterError(ErrAccountTypeWithPostings, fiber.StatusBadRequest, "cannot change type of account with postings")
	core.RegisterError(ErrAccountPostableWithPostings, fiber.StatusBadRequest, "cannot change postable of account with postings")
	core.RegisterError(ErrAccountDescendantHasPostings, fiber.StatusBadRequest, "cannot change type of account when descendant has postings")
	core.RegisterError(ErrAccountPostableWithChildren, fiber.StatusBadRequest, "cannot make account postable when it has children")
	core.RegisterError(ErrAccountDeleteWithChildren, fiber.StatusBadRequest, "cannot delete account with children; deactivate it instead")
	core.RegisterError(ErrAccountDeleteWithPostings, fiber.StatusBadRequest, "cannot delete account with postings; deactivate it instead")
	core.RegisterError(ErrAccountCodeExists, fiber.StatusConflict, "account code already exists in this parish")
	core.RegisterError(ErrParentAccountNotExist, fiber.StatusBadRequest, "parent account does not exist in this parish")
	core.RegisterError(ErrParentAccountPostable, fiber.StatusBadRequest, "parent account must not be postable")
	core.RegisterError(ErrAccountTypeMismatch, fiber.StatusBadRequest, "child account type must match parent account type")
	core.RegisterError(ErrAccountSelfParent, fiber.StatusBadRequest, "account cannot be its own parent")
	core.RegisterError(ErrAccountCycleDetected, fiber.StatusBadRequest, "cycle detected: account cannot have its descendant as parent")

	// Activate
	core.RegisterError(ErrParishNotFound, fiber.StatusBadRequest, "parish not found")
	core.RegisterError(ErrWorkspaceNotFound, fiber.StatusBadRequest, "workspace not found or does not belong to specified parish")

	// Report
	core.RegisterError(ErrAccountIdRequired, fiber.StatusBadRequest, "accountId is required")
	core.RegisterError(ErrAccountNotFoundInParish, fiber.StatusNotFound, "account not found in this parish")
	core.RegisterError(ErrDatesRequired, fiber.StatusBadRequest, "from and to dates are required")
	core.RegisterError(ErrInvalidFromDate, fiber.StatusBadRequest, "invalid from date")
	core.RegisterError(ErrInvalidToDate, fiber.StatusBadRequest, "invalid to date")
	core.RegisterError(ErrFromDateAfterToDate, fiber.StatusBadRequest, "from date must be before or equal to to date")
	core.RegisterError(ErrInvalidCursor, fiber.StatusBadRequest, "invalid cursor")
	core.RegisterError(ErrAsOfDateRequired, fiber.StatusBadRequest, "asOf date is required")
	core.RegisterError(ErrInvalidAsOfDate, fiber.StatusBadRequest, "invalid asOf date")
	core.RegisterError(ErrDateRangeExceedsLimit, fiber.StatusBadRequest, "date range exceeds 366 days")
	core.RegisterError(ErrParentAccountNotFound, fiber.StatusNotFound, "parent account not found in this parish")
	core.RegisterError(ErrParentMustBeHeader, fiber.StatusBadRequest, "parent account must be a header account")
	core.RegisterError(ErrInvalidAccountIDParam, fiber.StatusBadRequest, "invalid accountId")
	core.RegisterError(ErrInvalidParentAccountIDParam, fiber.StatusBadRequest, "invalid parentAccountId")

	// Transaction
	core.RegisterError(ErrInvalidDateFormat, fiber.StatusBadRequest, "invalid date format")
	core.RegisterError(ErrInvalidTxNumberFormat, fiber.StatusUnprocessableEntity, "invalid transaction number format")
	core.RegisterError(ErrInvalidPostingSide, fiber.StatusBadRequest, "invalid posting side: must be debit or credit")
	core.RegisterError(ErrTxMissingPostings, fiber.StatusBadRequest, "transaction must have at least one debit posting and one credit posting")
	core.RegisterError(ErrTxUnbalanced, fiber.StatusBadRequest, "transaction is unbalanced: sum of debits must equal sum of credits and be greater than zero")
	core.RegisterError(ErrAccountsNotFound, fiber.StatusBadRequest, "one or more accounts not found in this parish")
	core.RegisterError(ErrAccountNotActive, fiber.StatusBadRequest, "account is not active")
	core.RegisterError(ErrAccountIsHeader, fiber.StatusBadRequest, "account is a header account and cannot accept postings")
	core.RegisterError(ErrTransactionHasNoWorkspace, fiber.StatusUnprocessableEntity, "transaction has no workspace; attachments require a workspace")
	core.RegisterError(ErrTxNumberExists, fiber.StatusConflict, "transaction number already exists in this parish")
	core.RegisterError(ErrTxNumberAllocationFailed, fiber.StatusConflict, "could not allocate transaction number after retries")
	core.RegisterError(ErrTxNotPosted, fiber.StatusConflict, "transaction is not in posted status")
}
