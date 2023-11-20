import { createStore } from 'vuex'
import { MediaEnum } from '@/components/utils/utils'

export default createStore({
	state: {
		ScreenLevel: MediaEnum.Medium,
		IsMobile: true,
		ActivePage: '',
	},
	getters: {
		GetScreenLevel(state: any): typeof MediaEnum {
			return state.ScreenLevel
		},
		GetIsMobile(state: any): boolean {
			return state.ScreenLevel < 2
		},
		GetActivePage(state: any): String {
			return state.ActivePage
		},
	},
	mutations: {
		ScreenLevelMutation(state:any, level: typeof MediaEnum) {
			state.ScreenLevel = level;
			state.IsMobile = state.ScreenLevel < 2;
		},
		ActivePageMutation(state: any, page: String) {
			state.ActivePage = page;
		},
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
		},
		SetActivePage(state:any, page: String) {
			state.commit('ActivePageMutation', page);
		},
	}
})
