// Validates: REQ-USR.UI.gui-implementation
// F11.1 — public landing page renders hero, email CTA, audience sections,
// and clickable demo ontology cards without auth.

import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { VEDO_DEMOS } from "../data/demos";
import LandingPage from "../pages/LandingPage.vue";

describe("LandingPage", () => {
	it("should render hero title and positioning statement", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.find(".landing-hero__title").text()).toBe(
			"GitHub for ontologies",
		);
		expect(wrapper.find(".landing-hero__subtitle").exists()).toBe(true);
	});

	it("should render email CTA form with status on submit", async () => {
		const wrapper = mount(LandingPage);
		const input = wrapper.find(".landing-cta__input");
		await input.setValue("demo@company.com");
		await wrapper.find(".landing-cta__btn").trigger("submit");
		const status = wrapper.find(".landing-cta__status");
		expect(status.exists()).toBe(true);
		expect(status.text()).toContain("demo@company.com");
	});

	it("should render three audience cards", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.findAll(".audience-card").length).toBe(3);
	});

	it("should render demo ontology cards from VEDO_DEMOS", () => {
		const wrapper = mount(LandingPage);
		const cards = wrapper.findAll(".demo-card");
		expect(cards.length).toBe(VEDO_DEMOS.length);
	});

	it("should emit openDemo with the selected demo on card click", async () => {
		const wrapper = mount(LandingPage);
		const firstCard = wrapper.findAll(".demo-card")[0];
		await firstCard.trigger("click");
		const emitted = wrapper.emitted("openDemo");
		expect(emitted).toBeTruthy();
		expect((emitted?.[0]?.[0] as { id: string }).id).toBe(VEDO_DEMOS[0].id);
	});
});
