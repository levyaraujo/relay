package shared

type ContextKey string

const (
	UserID    ContextKey = "userID"
	CompanyID ContextKey = "companyID"
)