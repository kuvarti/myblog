package models

type UserModel struct{
	Username string `json:"Username"`
	Password string `json:"Username"`
	UserType string `json:"UserPrivilige"`
}

type LoginModel struct{
	Username string `json:"userName"`
	Password string `json:"passWord"`
}
