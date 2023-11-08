import { createStore } from 'vuex'
import { MediaEnum } from '@/components/utils/utils'

export default createStore({
	state: {
		ScreenLevel: MediaEnum.Medium,
		isMobile: true
	},
	getters: {
		GetScreenLevel(state: any): typeof MediaEnum {
			return state.ScreenLevel
		},
		GetIsMobile(state: any): boolean {
			return state.ScreenLevel < 2
		}
	},
	mutations: {
		ScreenLevelMutation(state:any, level: typeof MediaEnum) {
			state.ScreenLevel = level;
			state.isMobile = state.ScreenLevel < 2;
		}
	},
	actions: {
		SetScreenLevel(state:any, width: number) {
			if (width >= 1536) //* 1400 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.ExtraHigh);
			else if (width >= 1280) //* 1200 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.High);
			else if (width >= 1024) //* 992 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.Medium);
			else if (width >= 768) //* 768 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.Small);
			else //* 768 alti
				state.commit('ScreenLevelMutation', MediaEnum.ExtraSmall);
		}
	}
})
