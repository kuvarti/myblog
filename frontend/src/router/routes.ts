import sideMenu from "@/views/sideMenu.vue";
import sidePanel from "@/views/sidePanel.vue";
import contents from "@/views/contents.vue";
import lists from "@/views/lists.vue";
import panel from "@/views/panel.vue"

let pageArray = [sideMenu, sidePanel, contents, lists, panel];
export default pageArray;
export { default as sideMenu } from "@/views/sideMenu.vue";
export { default as sidePanel } from "@/views/sidePanel.vue";
export { default as contents } from "@/views/contents.vue";
export { default as lists} from "@/views/lists.vue";
export { default as panel} from "@/views/panel.vue";

