// Validates: REQ-USR.UI.gui-implementation
// F11.1 — public landing page per ADR-DES.UI.public-landing-architecture:
// header nav, hero (positioning + dual CTA), 5 demo cards, 5 role cards,
// trust/metrics, footer. No auth required.

import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { VEDO_DEMOS } from "../data/demos";
import LandingPage from "../pages/LandingPage.vue";

describe("LandingPage", () => {
	it("should render header with logo and login control", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.find(".landing-header__logo").text()).toBe("VEDO Hub");
		expect(wrapper.find(".landing-header__login").exists()).toBe(true);
	});

	it("should render hero title per ADR positioning", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.find(".landing-hero__title").text()).toBe(
			"VEDO Hub — GitHub + Hugging Face для онтологий",
		);
		expect(wrapper.findAll(".landing-hero__sub").length).toBe(3);
	});

	it("should render dual hero CTA (primary + outline)", () => {
		const wrapper = mount(LandingPage);
		const primary = wrapper.find(".btn--primary");
		const outline = wrapper.find(".btn--outline");
		expect(primary.exists()).toBe(true);
		expect(primary.text()).toContain("Создать аккаунт");
		expect(outline.exists()).toBe(true);
		expect(outline.text()).toContain("Смотреть демо");
	});

	it("should render five demo ontology cards from VEDO_DEMOS", () => {
		const wrapper = mount(LandingPage);
		const cards = wrapper.findAll(".demo-card");
		expect(cards.length).toBe(5);
		expect(cards.length).toBe(VEDO_DEMOS.length);
	});

	it("should render five role cards", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.findAll(".role-card").length).toBe(5);
	});

	it("should render trust metrics section with four metrics", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.findAll(".metric").length).toBe(4);
		expect(wrapper.find(".landing-open-core").exists()).toBe(true);
	});

	it("should render footer with columns and copyright", () => {
		const wrapper = mount(LandingPage);
		expect(wrapper.findAll(".footer-col").length).toBe(5);
		expect(wrapper.find(".landing-footer__bottom").text()).toContain(
			"© 2026 VEDO Hub",
		);
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
