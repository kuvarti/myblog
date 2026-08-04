import axios, { type AxiosInstance } from "axios";
import { type MenuListModal } from '@/types/MenuListModal'
import { type PageResponseModal } from '@/types/PageResponseModal'

export class serviceClass{
	protected apiClient:AxiosInstance;
	private userToken: string;
	constructor(path: string = "") {
		this.apiClient = axios.create({
			// Backend API base URL comes from VITE_API_BASE_URL, inlined by Vite at
			// build time (set it in .env.production or the build command). Falls
			// back to the local dev backend when unset.
			baseURL: (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api') + path,
			headers: {
				Accept: 'application/json',
				"Content-type": "application/json"
			}
		});
		this.userToken = "";
	}
	public async getMenu() : Promise<Array<MenuListModal>> {
		return new Promise((resolve) => {
			this.apiClient.get('MenuList/Menu').catch((reason) => {
				console.log('apiget field fail:', reason);
				resolve([])
			}).then((value) => {
				if (value && value.data)
					resolve(value.data)
				else
					resolve([])
			})
		})
	}//0.0.0.0:8080/Page?PageName=SoLong
	public async getPage(pageName: string) : Promise<PageResponseModal> {
		return new Promise((resolve) => {
			this.apiClient.get('Page?PageName=' + pageName).catch((reason) => {
				console.log('apiget field fail:', reason);
				resolve({Page: '', ViewType: ''})
			}).then((value) => {
				if (value && value.data){
					value.data.Page = value.data.Page.replace(/\/n/g, '\n').replace(/\\n/g, '\n')
					resolve(value.data)
				}
				else
					resolve({Page: '', ViewType: ''})
			})
		})
	}
	// Resolve a page by its route Path (the per-page routing key). Mirrors
	// getPage's newline handling; the leading "/" is encoded so the query stays
	// a single value.
	public async getPageByPath(path: string) : Promise<PageResponseModal> {
		return new Promise((resolve) => {
			this.apiClient.get('Page?Path=' + encodeURIComponent(path)).catch((reason) => {
				console.log('apiget field fail:', reason);
				resolve({Page: '', ViewType: ''})
			}).then((value) => {
				if (value && value.data){
					value.data.Page = value.data.Page.replace(/\/n/g, '\n').replace(/\\n/g, '\n')
					resolve(value.data)
				}
				else
					resolve({Page: '', ViewType: ''})
			})
		})
	}
};

export type ServiceType = serviceClass | undefined | null;
export type serviceClassImpl = serviceClass;
export default new serviceClass();
