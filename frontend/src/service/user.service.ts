import { type User} from '@/types/UserModal'
import process from 'process';
import bcrypt from "bcryptjs";
import { serviceClass } from '@/service/service'
import { ref } from 'vue';

class UserService extends serviceClass{
	private Token: string | null;
	// private User: User;
	constructor() {
		super("/ControlPanel/auth")
		this.Token = null
	}
	private async HashPasswd(pass: string): Promise<string> {
		return new Promise(async (resolve) => {
			const saltRounds = 10; // Go ile uyumlu olması için cost değeri
			const hash = await bcrypt.hash(pass, saltRounds);
			resolve(hash);
		})
	}
	public IsLogin = ref(false);
	public async Login(user: string, pass: string): Promise<boolean>{
		return new Promise(async (resolve) => {
			this.apiClient.post("/login", {
				"userName": user,
				"passWord": pass
			}).catch((reason) => {
				resolve(false)
			}).then((value)=> {
				console.log(value);
				if (value && value.data == "invalid Username and Password") {
					resolve(false)
				} else if (value && value.data.token){
					this.Token = value.data.token;
					console.log(this.Token);
					this.IsLogin.value = true;
					resolve(true)
				} else {
					resolve(false)
				}
			})
		})
	}
}

export type UserServiceType = UserService
export default new UserService();
