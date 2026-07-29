// Validates: REQ-USR.UI.import-export
// Validates: REQ-USR.UI.doc-extract-upload
// Validates: REQ-USR.UI.excel-error-report
// Validates: REQ-USR.UI.excel-progress
// Validates: REQ-USR.UI.excel-templates
// Validates: REQ-USR.UI.excel-preview
// Validates: REQ-USR.UI.query-export-availability
// Validates: REQ-USR.UI.query-export-localization
// Validates: REQ-USR.UI.query-export-notifications
//
// P0 placeholder specs for import/export and document extraction UI.
// Replace it.todo with real implementations when corresponding UI components are built.

import { describe, it } from "vitest";

describe.skip("Import/Export (REQ-USR.UI.import-export)", () => {
	it.todo("should provide a drag-and-drop import zone for ontology files");
	it.todo("should support export in Turtle, JSON-LD, and RDF/XML formats");
	it.todo("should show import progress with file name and size");
});

describe.skip("Document Extract Upload (REQ-USR.UI.doc-extract-upload)", () => {
	it.todo("should accept document uploads for AI extraction via drag-and-drop");
	it.todo("should show supported file formats on the upload area");
	it.todo("should validate file size before upload and show error if exceeded");
});

describe.skip("Excel Error Report (REQ-USR.UI.excel-error-report)", () => {
	it.todo(
		"should display import errors in a structured table with row references",
	);
});

describe.skip("Excel Progress (REQ-USR.UI.excel-progress)", () => {
	it.todo("should show import progress when processing large Excel files");
});

describe.skip("Excel Templates (REQ-USR.UI.excel-templates)", () => {
	it.todo("should provide downloadable Excel templates with required columns");
});

describe.skip("Excel Preview (REQ-USR.UI.excel-preview)", () => {
	it.todo("should show a column mapping preview before importing Excel files");
	it.todo("should auto-detect column-to-field mapping from header names");
	it.todo("should allow manual correction of column mapping");
	it.todo("should display the first 3-5 data rows as examples");
	it.todo("should update preview dynamically when mapping changes");
});

describe.skip("Query Export Availability (REQ-USR.UI.query-export-availability)", () => {
	it.todo("should indicate when query export is ready for download");
});

describe.skip("Query Export Localization (REQ-USR.UI.query-export-localization)", () => {
	it.todo("should localize export format names and descriptions");
});

describe.skip("Query Export Notifications (REQ-USR.UI.query-export-notifications)", () => {
	it.todo("should notify the user when a query export completes");
});
