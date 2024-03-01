import axios, { type AxiosInstance } from "axios";
import { type MenuListModal } from '@/types/MenuListModal'
import { type PageResponseModal } from '@/types/PageResponseModal'

export class serviceClass{
	protected apiClient:AxiosInstance;
	private userToken: string;
	constructor(path: string = "") {
		this.apiClient = axios.create({
			baseURL: 'http://localhost:8080/api' + path,
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
};

export type ServiceType = serviceClass | undefined | null;
export type serviceClassImpl = serviceClass;
export default new serviceClass();
