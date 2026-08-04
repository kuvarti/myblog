package models

type PageSummary struct {
	PageName string `json:"PageName" bson:"PageName"`
	Path     string `json:"Path" bson:"Path"`
	ViewType string `json:"ViewType" bson:"ViewType"`
	Order    int    `json:"Order" bson:"Order"`
	Hidden   bool   `json:"-" bson:"Hidden"`
	Visible  bool   `json:"Visible" bson:"-"`
}

type MenuBinding struct {
	Name    string `json:"Name"`
	Caption string `json:"Caption"`
}

type PageDetail struct {
	PageName string       `json:"PageName"`
	Path     string       `json:"Path"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type CreatePageRequest struct {
	PageName string       `json:"PageName"`
	Path     string       `json:"Path"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type UpdatePageRequest struct {
	Path     string       `json:"Path"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type PreviewRequest struct {
	Source string `json:"Source"`
}

type ReorderRequest struct {
	PageNames []string `json:"PageNames"`
}

type VisibilityRequest struct {
	PageName string `json:"PageName"`
	Visible  bool   `json:"Visible"`
}
