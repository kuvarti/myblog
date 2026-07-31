package models

type UserModel struct{
	Username string `json:"Username"`
	Password string `json:"Password"`
	UserType string `json:"UserPrivilige"`
}

type LoginModel struct{
	Username string `json:"userName"`
	Password string `json:"passWord"`
}
