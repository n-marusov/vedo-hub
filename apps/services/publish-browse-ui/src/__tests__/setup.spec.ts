// Validates: REQ-USR.UI.gui-implementation
//
// Smoke test for the publish-browse-ui stub application.

import { describe, expect, it } from "vitest";

describe("publish-browse-ui", () => {
	it("should have a main entry point", async () => {
		// The main module creates a Vue app — verify it can be imported
		// without throwing at compile time.
		const main = await import("../main");
		expect(main).toBeDefined();
	});
});
