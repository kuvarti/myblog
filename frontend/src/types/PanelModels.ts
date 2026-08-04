export interface MenuBinding {
	Name: string
	Caption: string
}

export interface PanelPageSummary {
	PageName: string
	Path: string
	ViewType: string
	Order: number
	Visible: boolean
}

export interface PanelPageDetail {
	PageName: string
	Path: string
	Source: string
	ViewType: string
	Tags: string[]
	Summary: string
	Image: string
	ListTags: string[]
	Menu: MenuBinding | null
}
