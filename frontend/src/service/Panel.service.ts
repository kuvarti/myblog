import { serviceClass } from '@/service/BaseAPI.service'
import { LocalStorageService } from '@/service/LocalStorage.service'
import UserService from '@/service/User.service'
import type { PanelPageSummary, PanelPageDetail, MenuBinding } from '@/types/PanelModels'

export interface SavePagePayload {
	PageName: string
	Path: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}

class PanelService extends serviceClass {
	private localStorage: LocalStorageService
	constructor() {
		super('/auth/ControlPanel')
		this.localStorage = new LocalStorageService()
	}
	private authConfig() {
		const token = this.localStorage.GetData('AuthToken')
		return { headers: { Authorization: `Bearer ${token ? token : 'nullvalue'}` } }
	}
	private handleAuthError(err: any): never {
		if (err?.response?.status === 401) {
			UserService.IsLogin.value = false
		}
		throw err
	}
	public async listPages(): Promise<PanelPageSummary[]> {
		return this.apiClient.get('/Pages', this.authConfig())
			.then((r) => r.data as PanelPageSummary[])
			.catch((e) => { try { this.handleAuthError(e) } catch { /* swallow */ } return [] })
	}
	public async getPageDetail(name: string): Promise<PanelPageDetail> {
		return this.apiClient.get(`/Pages/${encodeURIComponent(name)}`, this.authConfig())
			.then((r) => r.data as PanelPageDetail)
			.catch((e) => this.handleAuthError(e))
	}
	public async createPage(payload: SavePagePayload): Promise<void> {
		return this.apiClient.post('/Pages', payload, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async updatePage(name: string, payload: Omit<SavePagePayload, 'PageName'>): Promise<void> {
		return this.apiClient.put(`/Pages/${encodeURIComponent(name)}`, payload, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async deletePage(name: string): Promise<void> {
		return this.apiClient.delete(`/Pages/${encodeURIComponent(name)}`, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async reorderPages(pageNames: string[]): Promise<void> {
		return this.apiClient.put('/PageOrder', { PageNames: pageNames }, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async setVisibility(pageName: string, visible: boolean): Promise<void> {
		return this.apiClient.put('/PageVisibility', { PageName: pageName, Visible: visible }, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async preview(source: string): Promise<string> {
		return this.apiClient.post('/Preview', { Source: source }, this.authConfig())
			.then((r) => r.data.Html as string)
			.catch((e) => { try { this.handleAuthError(e) } catch { /* swallow */ } return '' })
	}
}

export type PanelServiceType = PanelService
export default new PanelService()
