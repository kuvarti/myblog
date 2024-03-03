export class LocalStorageService {
	constructor() {

	}

	public SaveData(key: string, value: string) {
		localStorage.setItem(key, value);
	}

	public GetData(key: string): string | null {
		let response = localStorage.getItem(key);
		return response;
	}
}
