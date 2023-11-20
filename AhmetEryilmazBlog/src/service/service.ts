import axios, { type AxiosInstance } from "axios";

class serviceClass{
	private apiClient:AxiosInstance;
	constructor() {
		console.log('Constructor called')
		this.apiClient = axios.create({
			baseURL: 'http://localhost:8080',
			headers: {
				Accept: 'application/json',
				"Content-type": "application/json"
			}
		});
	}
	public async getMenu() : Promise<Array<[string, string]>> {
		return new Promise((resolve) => {
			this.apiClient.get('MenuList').catch((reason) => {
				console.log('apiget field fail:', reason);
				resolve([])
			}).then((value) => {
				if (value && value.data)
					resolve(value.data.menu)
				else
					resolve([])
			})
		})
	}
};

export type ServiceType = serviceClass | undefined | null;
export default new serviceClass();
