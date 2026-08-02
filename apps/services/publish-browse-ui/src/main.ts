import { computed, createApp, h, ref } from "vue";
import { type DemoOntology, findDemo } from "./data/demos";
import DemoOntologyView from "./pages/DemoOntologyView.vue";
import LandingPage from "./pages/LandingPage.vue";

// Lightweight hash router for the public landing (F11.1): #/ shows the landing
// page, #/demo/<id> shows a demo ontology view. No vue-router dependency — the
// public showcase is a two-view surface.
const currentHash = ref(window.location.hash || "#/");

function parseHash(): { view: "landing" } | { view: "demo"; id: string } {
	const hash = currentHash.value;
	const match = hash.match(/^#\/demo\/(.+)$/);
	if (match) return { view: "demo", id: decodeURIComponent(match[1]) };
	return { view: "landing" };
}

const route = computed(() => parseHash());
const activeDemo = computed<DemoOntology | null>(() =>
	route.value.view === "demo" ? (findDemo(route.value.id) ?? null) : null,
);

window.addEventListener("hashchange", () => {
	currentHash.value = window.location.hash || "#/";
});

function openDemo(demo: DemoOntology) {
	window.location.hash = `/demo/${encodeURIComponent(demo.id)}`;
}

function goBack() {
	window.location.hash = "#/";
}

createApp({
	render() {
		if (route.value.view === "demo") {
			return h(DemoOntologyView, { demo: activeDemo.value, onBack: goBack });
		}
		return h(LandingPage, { onOpenDemo: openDemo });
	},
}).mount("#app");
