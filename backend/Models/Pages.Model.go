package models

type PageModel struct {
	PageName	string	`json:"PageName" gorm:"unique"`
	Page		string	`json:"Page"`
	Hash		[]byte	`json:"Hash"`
	Text		string	`json:"Text"`
}

type PageResponse struct {
	PageName	string	`json:"PageName"`
	Test		string	`json:"Page"`
}
