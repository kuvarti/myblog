import { createStore } from 'vuex'
import { MediaEnum } from '@/components/utils/utils'

export default createStore({
	state: {
		ScreenLevel: MediaEnum.Medium
	},
	getters: {
		GetScreenLevel(state: any): typeof MediaEnum {
			return state.ScreenLevel
		}
	},
	mutations: {
		ScreenLevelMutation(state:any, level: typeof MediaEnum) {
			state.ScreenLevel = level;
		}
	},
	actions: {
		SetScreenLevel(state:any, width: number) {
			if (width < 576) {
				state.commit('ScreenLevelMutation', MediaEnum.ExtraSmall);
			} else if (width < 768) {
				state.commit('ScreenLevelMutation', MediaEnum.Small);
			} else if (width < 992) {
				state.commit('ScreenLevelMutation', MediaEnum.Medium);
			} else if (width < 1200) {
				state.commit('ScreenLevelMutation', MediaEnum.High);
			} else {
				state.commit('ScreenLevelMutation', MediaEnum.ExtraHigh);
			}
		}
	}
})
