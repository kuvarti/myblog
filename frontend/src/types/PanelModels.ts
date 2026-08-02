export interface MenuBinding {
	Name: string
	Caption: string
	Path: string
}

export interface PanelPageSummary {
	PageName: string
	ViewType: string
}

export interface PanelPageDetail {
	PageName: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}
