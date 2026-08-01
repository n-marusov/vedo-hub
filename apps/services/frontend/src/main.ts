// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Frontend entry point — Apollo Client, Vue Router, error presentation, i18n initialization

import { ApolloClients } from "@vue/apollo-composable";
import { createApp, h, provide } from "vue";
import App from "./App.vue";
import { apolloClient } from "./apollo/client";
import { initSession } from "./auth/session";
import { useErrorPresentation } from "./composables/useErrorPresentation";
import { useI18n } from "./composables/useI18n";
import router from "./router";

// Styles
import "./styles/tokens.css";
import "./styles/global.css";
import "./styles/customer-overrides.css";

// Seed the auth session before any API client can fire a request — in
// SKIP_AUTH (dev/test) mode this writes the dev-minted JWT to
// localStorage.vedo-jwt-token so Apollo/Axios interceptors send a valid
// Bearer token (otherwise org write endpoints return 401).
initSession();

const app = createApp({
	setup() {
		provide(ApolloClients, { default: apolloClient });

		const errorPresentation = useErrorPresentation();
		errorPresentation.installGlobalHandler();

		const i18n = useI18n();
		i18n.setLocale("ru");

		return () => h(App);
	},
});

app.use(router);
app.mount("#app");
