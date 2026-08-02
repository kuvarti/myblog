package models

type PageSummary struct {
	PageName string `json:"PageName" bson:"PageName"`
	ViewType string `json:"ViewType" bson:"ViewType"`
}

type MenuBinding struct {
	Name    string `json:"Name"`
	Caption string `json:"Caption"`
	Path    string `json:"Path"`
}

type PageDetail struct {
	PageName string       `json:"PageName"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type CreatePageRequest struct {
	PageName string       `json:"PageName"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type UpdatePageRequest struct {
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type PreviewRequest struct {
	Source string `json:"Source"`
}
