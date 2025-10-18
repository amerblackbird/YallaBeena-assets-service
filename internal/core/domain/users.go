package domain

type UserType string

const (
	UserTypeCustomer UserType = "customer"
	UserTypeDriver   UserType = "driver"
	UserTypeAdmin    UserType = "admin"
)
