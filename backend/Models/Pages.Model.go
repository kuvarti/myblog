package models

type PageModel struct {
	PageName	string	`json:"PageName" gorm:"unique"`
	Page		string	`json:"Page"`
	Hash		[]byte	`json:"Hash"`
	Text		string	`json:"Text"`
	ViewType	string	`json:"ViewType"`
	Order		int	`json:"Order"`
	Hidden		bool	`json:"-" bson:"Hidden"`
}
