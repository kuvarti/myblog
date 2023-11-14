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
	public async getMenu() {
		return new Promise((resolve) => {
			this.apiClient.get('someEndpoint').catch((reason) => {
				console.log('apiget field fail:', reason);
			}).then((value) => {
				if (value)
				{
					resolve(value.data.message)
				}
			})
		})
	}
};

export type ServiceType = serviceClass | undefined;
export default new serviceClass();
