export interface MenuBinding {
	Name: string
	Caption: string
	Path: string
}

export interface PanelPageSummary {
	PageName: string
	ViewType: string
	Order: number
}

export interface PanelPageDetail {
	PageName: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}
