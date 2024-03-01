import sideMenu from "@/views/sideMenu.vue";
import viewProfile from "@/views/viewProfile.vue";
import contents from "@/views/contents.vue";
import lists from "@/views/lists.vue";
import panel from "@/views/panel.vue"

let pageArray = [sideMenu, viewProfile, contents, lists, panel];
export default pageArray;
export { default as sideMenu } from "@/views/sideMenu.vue";
export { default as viewProfile } from "@/views/viewProfile.vue";
export { default as contents } from "@/views/contents.vue";
export { default as lists} from "@/views/lists.vue";
export { default as panel} from "@/views/panel.vue";

