package finance

import "github.com/paroki/domus/api/internal/modules/finance/model"

var (
	// Context
	ErrContextNotActive = model.ErrContextNotActive

	// Account
	ErrAccountTypeWithPostings      = model.ErrAccountTypeWithPostings
	ErrAccountPostableWithPostings  = model.ErrAccountPostableWithPostings
	ErrAccountDescendantHasPostings = model.ErrAccountDescendantHasPostings
	ErrAccountPostableWithChildren  = model.ErrAccountPostableWithChildren
	ErrAccountDeleteWithChildren    = model.ErrAccountDeleteWithChildren
	ErrAccountDeleteWithPostings    = model.ErrAccountDeleteWithPostings
	ErrAccountCodeExists            = model.ErrAccountCodeExists
	ErrParentAccountNotExist        = model.ErrParentAccountNotExist
	ErrParentAccountPostable        = model.ErrParentAccountPostable
	ErrAccountTypeMismatch          = model.ErrAccountTypeMismatch
	ErrAccountSelfParent            = model.ErrAccountSelfParent
	ErrAccountCycleDetected         = model.ErrAccountCycleDetected

	// Activate
	ErrParishNotFound    = model.ErrParishNotFound
	ErrWorkspaceNotFound = model.ErrWorkspaceNotFound

	// Report
	ErrAccountIdRequired           = model.ErrAccountIdRequired
	ErrAccountNotFoundInParish     = model.ErrAccountNotFoundInParish
	ErrDatesRequired               = model.ErrDatesRequired
	ErrInvalidFromDate             = model.ErrInvalidFromDate
	ErrInvalidToDate               = model.ErrInvalidToDate
	ErrFromDateAfterToDate         = model.ErrFromDateAfterToDate
	ErrInvalidCursor               = model.ErrInvalidCursor
	ErrAsOfDateRequired            = model.ErrAsOfDateRequired
	ErrInvalidAsOfDate             = model.ErrInvalidAsOfDate
	ErrDateRangeExceedsLimit       = model.ErrDateRangeExceedsLimit
	ErrParentAccountNotFound       = model.ErrParentAccountNotFound
	ErrParentMustBeHeader          = model.ErrParentMustBeHeader
	ErrInvalidAccountIDParam       = model.ErrInvalidAccountIDParam
	ErrInvalidParentAccountIDParam = model.ErrInvalidParentAccountIDParam

	// Transaction
	ErrInvalidDateFormat         = model.ErrInvalidDateFormat
	ErrInvalidTxNumberFormat     = model.ErrInvalidTxNumberFormat
	ErrInvalidPostingSide        = model.ErrInvalidPostingSide
	ErrTxMissingPostings         = model.ErrTxMissingPostings
	ErrTxUnbalanced              = model.ErrTxUnbalanced
	ErrAccountsNotFound          = model.ErrAccountsNotFound
	ErrAccountNotActive          = model.ErrAccountNotActive
	ErrAccountIsHeader           = model.ErrAccountIsHeader
	ErrTransactionHasNoWorkspace = model.ErrTransactionHasNoWorkspace
	ErrTxNumberExists            = model.ErrTxNumberExists
	ErrTxNumberAllocationFailed  = model.ErrTxNumberAllocationFailed
	ErrTxNotPosted               = model.ErrTxNotPosted
)
