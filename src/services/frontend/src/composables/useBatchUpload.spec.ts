// Unit tests for useBatchUpload composable
// Tests: multi-file upload, deduplication logic, conflict resolution, partial failure handling

import { beforeEach, describe, expect, it, vi } from "vitest";
import { useBatchUpload } from "./useBatchUpload";

// Mock default — all uploads succeed with 1 step each
vi.mock("../api/extraction", () => ({
	batchUploadDocuments: vi
		.fn()
		.mockImplementation(async (opts: { files: File[] }) => {
			return opts.files.map((file: File) => ({
				id: `preview-${file.name}`,
				ontologyId: "test-onto",
				sourceFile: file.name,
				steps: [
					{
						id: `step-${file.name}-1`,
						operation: "CREATE_CLASS",
						entityId: "Person",
						label: "Person",
						included: true,
					},
				],
				totalSteps: 1,
				createdAt: "2026-07-16T12:00:00Z",
			}));
		}),
	ALLOWED_FORMATS: [
		".md",
		".txt",
		".pdf",
		".docx",
		".json",
		".xml",
		".csv",
		".xlsx",
	],
	FORMAT_LABELS: {
		".md": "Markdown",
		".txt": "Text",
		".pdf": "PDF",
		".docx": "DOCX",
		".json": "JSON",
		".xml": "XML",
		".csv": "CSV",
		".xlsx": "XLSX",
	},
}));

function createMockFile(name: string): File {
	return new File(["test content"], name, { type: "text/plain" });
}

describe("useBatchUpload", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe("Initial state", () => {
		it("starts with empty files and no steps", () => {
			const { files, mergedSteps, isProcessing } = useBatchUpload();

			expect(files.value).toEqual([]);
			expect(mergedSteps.value).toEqual([]);
			expect(isProcessing.value).toBe(false);
		});

		it("has zero counts initially", () => {
			const {
				pendingCount,
				uploadingCount,
				successCount,
				failedCount,
				totalCount,
			} = useBatchUpload();

			expect(pendingCount.value).toBe(0);
			expect(uploadingCount.value).toBe(0);
			expect(successCount.value).toBe(0);
			expect(failedCount.value).toBe(0);
			expect(totalCount.value).toBe(0);
		});
	});

	describe("Adding files", () => {
		it("adds files to the queue", () => {
			const { files, addFiles } = useBatchUpload();

			addFiles([createMockFile("test1.md"), createMockFile("test2.txt")]);

			expect(files.value.length).toBe(2);
			expect(files.value[0].fileName).toBe("test1.md");
			expect(files.value[1].fileName).toBe("test2.txt");
			expect(files.value[0].status).toBe("pending");
		});

		it("skips duplicate filenames", () => {
			const { files, addFiles } = useBatchUpload();

			addFiles([createMockFile("test.md")]);
			addFiles([createMockFile("test.md")]);

			expect(files.value.length).toBe(1);
		});

		it("tracks per-file status", () => {
			const { files, addFiles } = useBatchUpload();

			addFiles([createMockFile("doc1.md"), createMockFile("doc2.txt")]);

			expect(files.value[0].status).toBe("pending");
			expect(files.value[1].status).toBe("pending");
		});
	});

	describe("Removing files", () => {
		it("removes a file by name", () => {
			const { files, addFiles, removeFile } = useBatchUpload();

			addFiles([createMockFile("keep.md"), createMockFile("remove.txt")]);
			removeFile("remove.txt");

			expect(files.value.length).toBe(1);
			expect(files.value[0].fileName).toBe("keep.md");
		});
	});

	describe("Upload all files", () => {
		it("uploads all pending files", async () => {
			const { files, addFiles, uploadAll, allDone, successCount, mergedSteps } =
				useBatchUpload();

			addFiles([createMockFile("doc1.md"), createMockFile("doc2.txt")]);
			await uploadAll("test-onto");

			expect(allDone.value).toBe(true);
			expect(successCount.value).toBe(2);
			expect(files.value.every((f) => f.status === "success")).toBe(true);
			expect(mergedSteps.value.length).toBeGreaterThanOrEqual(2);
		});

		it("handles partial failure", async () => {
			const { batchUploadDocuments } = await import("../api/extraction");
			const mockBatch = batchUploadDocuments as ReturnType<typeof vi.fn>;
			mockBatch.mockImplementation(async (opts: { files: File[] }) => {
				if (opts.files[0].name === "fail.txt") {
					throw new Error("Upload failed");
				}
				return [
					{
						id: `preview-${opts.files[0].name}`,
						ontologyId: "test-onto",
						sourceFile: opts.files[0].name,
						steps: [],
						totalSteps: 0,
						createdAt: "2026-07-16T12:00:00Z",
					},
				];
			});

			const {
				files,
				addFiles,
				uploadAll,
				hasFailedFiles,
				successCount,
				failedCount,
			} = useBatchUpload();

			addFiles([createMockFile("good.md"), createMockFile("fail.txt")]);
			await uploadAll("test-onto");

			expect(hasFailedFiles.value).toBe(true);
			expect(successCount.value).toBe(1);
			expect(failedCount.value).toBe(1);
			// good.md succeeded
			expect(files.value.find((f) => f.fileName === "good.md")?.status).toBe(
				"success",
			);
			// fail.txt failed
			expect(files.value.find((f) => f.fileName === "fail.txt")?.status).toBe(
				"error",
			);
		});
	});

	describe("Deduplication", () => {
		it("detects duplicate entity IDs across files", async () => {
			const { batchUploadDocuments } = await import("../api/extraction");
			const mockBatch = batchUploadDocuments as ReturnType<typeof vi.fn>;
			let callCount = 0;
			mockBatch.mockImplementation(async (opts: { files: File[] }) => {
				callCount++;
				if (callCount === 1) {
					// File 1: Person + Organization
					return [
						{
							id: "preview-file1",
							ontologyId: "test-onto",
							sourceFile: opts.files[0].name,
							steps: [
								{
									id: "s1",
									operation: "CREATE_CLASS",
									entityId: "Person",
									label: "Person",
									included: true,
								},
								{
									id: "s2",
									operation: "CREATE_CLASS",
									entityId: "Organization",
									label: "Organization",
									included: true,
								},
							],
							totalSteps: 2,
							createdAt: "2026-07-16T12:00:00Z",
						},
					];
				}
				// File 2: Person (duplicate) + hasName
				return [
					{
						id: "preview-file2",
						ontologyId: "test-onto",
						sourceFile: opts.files[0].name,
						steps: [
							{
								id: "s3",
								operation: "CREATE_CLASS",
								entityId: "Person",
								label: "Person",
								included: true,
							},
							{
								id: "s4",
								operation: "CREATE_PROPERTY",
								entityId: "hasName",
								label: "hasName",
								included: true,
							},
						],
						totalSteps: 2,
						createdAt: "2026-07-16T12:00:00Z",
					},
				];
			});

			const { addFiles, uploadAll, mergedSteps } = useBatchUpload();

			addFiles([createMockFile("file1.md"), createMockFile("file2.md")]);
			await uploadAll("test-onto");

			// Should have 4 steps total (2 + 2)
			expect(mergedSteps.value.length).toBe(4);

			// Person from second file should be marked as duplicate
			const duplicates = mergedSteps.value.filter((s) => s.isDuplicate);
			// At least Person (entityId appears in both files) should be duplicate
			expect(duplicates.length).toBeGreaterThanOrEqual(1);
			expect(duplicates.some((d) => d.entityId === "Person")).toBe(true);
		});
	});

	describe("Conflict detection", () => {
		it("detects label conflicts between files", async () => {
			const { batchUploadDocuments } = await import("../api/extraction");
			const mockBatch = batchUploadDocuments as ReturnType<typeof vi.fn>;
			let callCount = 0;
			mockBatch.mockImplementation(async (opts: { files: File[] }) => {
				callCount++;
				if (callCount === 1) {
					// File 1: Person with label "Person"
					return [
						{
							id: "preview-f1",
							ontologyId: "test-onto",
							sourceFile: opts.files[0].name,
							steps: [
								{
									id: "s1",
									operation: "CREATE_CLASS",
									entityId: "Person",
									label: "Person",
									included: true,
								},
							],
							totalSteps: 1,
							createdAt: "2026-07-16T12:00:00Z",
						},
					];
				}
				// File 2: Same Person but label "Individual"
				return [
					{
						id: "preview-f2",
						ontologyId: "test-onto",
						sourceFile: opts.files[0].name,
						steps: [
							{
								id: "s2",
								operation: "CREATE_CLASS",
								entityId: "Person",
								label: "Individual",
								included: true,
							},
						],
						totalSteps: 1,
						createdAt: "2026-07-16T12:00:00Z",
					},
				];
			});

			const { addFiles, uploadAll, hasConflicts, conflicts } = useBatchUpload();

			addFiles([createMockFile("file1.md"), createMockFile("file2.md")]);
			await uploadAll("test-onto");

			// Should detect conflict for Person entity (different labels)
			expect(hasConflicts.value).toBe(true);
			expect(conflicts.value.length).toBeGreaterThanOrEqual(1);
			// The conflict should be about Person
			expect(conflicts.value[0].stepA.entityId).toBe("Person");
		});
	});

	describe("Reset", () => {
		it("resets all state", () => {
			const { files, addFiles, reset, totalCount } = useBatchUpload();

			addFiles([createMockFile("doc.md")]);
			expect(totalCount.value).toBe(1);

			reset();
			expect(files.value).toEqual([]);
			expect(totalCount.value).toBe(0);
		});
	});
});
