import { ref } from 'vue'

// A shared reactive "version" that the public side menu watches so it can
// refetch on demand. Panel actions that change what the nav shows (e.g. a
// visibility toggle) call refreshMenu() to bump it; sideMenu.vue reads
// menuVersion inside its watchEffect, so the bump triggers a re-fetch.
const menuVersion = ref(0)

export function useMenuVersion() {
	return menuVersion
}

export function refreshMenu(): void {
	menuVersion.value++
}
