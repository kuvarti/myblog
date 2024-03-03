import { type User} from '@/types/UserModal'
import process from 'process';
import bcrypt from "bcryptjs";
import { serviceClass } from '@/service/BaseAPI.service'
import { ref } from 'vue';
import { LocalStorageService } from '@/service/LocalStorage.service';

class UserService extends serviceClass{
	private localStorage: LocalStorageService;
	private Token: string | null;
	public IsLogin = ref(false);
	constructor() {
		super("/auth")
		this.localStorage = new LocalStorageService();
		this.Token = this.localStorage.GetData("AuthToken");
		this.AmIAuth();
	}
	private async HashPasswd(pass: string): Promise<string> {
		return new Promise(async (resolve) => {
			const saltRounds = 10; // Go ile uyumlu olması için cost değeri
			const hash = await bcrypt.hash(pass, saltRounds);
			resolve(hash);
		})
	}
	private AmIAuth() {
		this.apiClient.get("/AmIAuth", {
			headers: {
				"Authorization": `Bearer ${this.Token ? this.Token : "nullvalue"}`
			}
		}).then((value) => {
			let response:boolean = value?.data !== "true";
			this.IsLogin.value = response;
		}).catch(() => {
			this.IsLogin.value = false;
		})
	}
	public async Login(user: string, pass: string) {
		this.apiClient.post("/login", {
			"userName": user,
			"passWord": pass
		}).catch((reason) => {
			console.log(reason)
		}).then((value)=> {
			if (value?.data == "invalid Username and Password") {
				return ;
			} else if (value?.data?.token){
				this.Token = value.data.token;
				if (this.Token)
					this.localStorage.SaveData("AuthToken", this.Token);
			}
			this.AmIAuth();
		})
	}
}

export type UserServiceType = UserService
export default new UserService();
