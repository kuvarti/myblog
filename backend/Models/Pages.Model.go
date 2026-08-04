package models

type PageModel struct {
	PageName string   `json:"PageName" gorm:"unique"`
	Path     string   `json:"Path" bson:"Path"`
	Page     string   `json:"Page"`
	Hash     []byte   `json:"Hash"`
	Text     string   `json:"Text"`
	ViewType string   `json:"ViewType"`
	Order    int      `json:"Order"`
	Hidden   bool     `json:"-" bson:"Hidden"`
	Tags     []string `json:"Tags" bson:"Tags"`
	Summary  string   `json:"Summary" bson:"Summary"`
	Image    string   `json:"Image" bson:"Image"`
	ListTags []string `json:"ListTags" bson:"ListTags"`
}
