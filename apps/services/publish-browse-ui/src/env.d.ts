/// <reference types="vite/client" />

// Vue SFC module declaration so plain `tsc --noEmit` can resolve .vue imports.
declare module "*.vue" {
	import type { DefineComponent } from "vue";
	const component: DefineComponent<
		Record<string, unknown>,
		Record<string, unknown>,
		unknown
	>;
	export default component;
}
