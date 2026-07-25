#!/usr/bin/env node
// @ctx: Auto-fix TypeScript errors in Vue SFC files
// Usage: pnpm fix:types
// Fixes: _var → var (when used in template), :name → :displayName (Avatar), :label → :text (Badge), null → undefined, as const for literals

import { execSync } from "node:child_process";
import * as fs from "node:fs";
import * as path from "node:path";

const SRC_DIR = path.resolve("src");

function runVueTsc() {
	try {
		execSync("npx vue-tsc --noEmit", { cwd: SRC_DIR, stdio: "pipe" });
		return [];
	} catch (e) {
		const output = e.stdout?.toString() + e.stderr?.toString();
		return parseErrors(output);
	}
}

function parseErrors(output) {
	const errors = [];
	const regex = /^(.+?)\((\d+),\d+\):\s+error\s+TS(\d+):\s+(.+)$/gm;
	let match = regex.exec(output);
	while (match !== null) {
		errors.push({
			file: match[1],
			line: Number(match[2]),
			code: match[3],
			message: match[4],
		});
		match = regex.exec(output);
	}
	return errors;
}

function fixFile(filePath, fixes) {
	const content = fs.readFileSync(filePath, "utf-8");
	const lines = content.split("\n");
	let changed = false;
	for (const fix of fixes) {
		const idx = fix.line - 1;
		if (idx >= 0 && idx < lines.length && lines[idx].includes(fix.old)) {
			lines[idx] = lines[idx].replace(fix.old, fix.new);
			changed = true;
		}
	}
	if (changed) {
		fs.writeFileSync(filePath, lines.join("\n"), "utf-8");
		console.log(`  ✓ ${path.relative(SRC_DIR, filePath)}`);
	}
}

function applyFixes(errors) {
	const fixesByFile = new Map();
	for (const err of errors) {
		const fileKey = path.resolve(SRC_DIR, err.file);
		const content = fs.readFileSync(fileKey, "utf-8");
		const lines = content.split("\n");
		const line = lines[err.line - 1] || "";

		// TS6133: unused _var that's actually used in template → remove underscore
		if (err.code === "6133") {
			const varName =
				err.message.match(/"([^"]+)"/)?.[1] ||
				err.message.match(/'([^']+)'/)?.[1];
			if (varName?.startsWith("_")) {
				const clean = varName.slice(1);
				if (
					content.includes(`{{ ${clean}`) ||
					content.includes(`:${clean}`) ||
					content.includes(`v-for="${clean}`) ||
					content.includes(`v-if="${clean}`) ||
					content.includes(`v-model="${clean}`) ||
					content.includes(`v-show="${clean}`) ||
					content.includes(`v-text="${clean}`) ||
					content.includes(`v-html="${clean}`)
				) {
					const fixes = fixesByFile.get(fileKey) || [];
					fixes.push({
						line: err.line,
						old: `const ${varName}`,
						new: `const ${clean}`,
					});
					fixesByFile.set(fileKey, fixes);
				}
			}
		}

		// TS2339: property doesn't exist → check for _prefixed version
		if (err.code === "2339") {
			const propName =
				err.message.match(/"([^"]+)"/)?.[1] ||
				err.message.match(/'([^']+)'/)?.[1];
			if (propName) {
				for (let i = 0; i < lines.length; i++) {
					if (
						lines[i].includes(`const _${propName}`) ||
						lines[i].includes(`const _${propName} =`)
					) {
						const fixes = fixesByFile.get(fileKey) || [];
						fixes.push({
							line: i + 1,
							old: `const _${propName}`,
							new: `const ${propName}`,
						});
						fixesByFile.set(fileKey, fixes);
						break;
					}
				}
			}
		}

		// TS2345: Avatar :name → :displayName
		if (
			err.code === "2345" &&
			(err.message.includes("displayName") || err.message.includes("Avatar"))
		) {
			const fixes = fixesByFile.get(fileKey) || [];
			if (line.includes(":name="))
				fixes.push({ line: err.line, old: ":name=", new: ":displayName=" });
			fixesByFile.set(fileKey, fixes);
		}

		// TS2345: Badge :label → :text
		if (err.code === "2345" && err.message.includes("Badge")) {
			const fixes = fixesByFile.get(fileKey) || [];
			if (line.includes(":label="))
				fixes.push({ line: err.line, old: ":label=", new: ":text=" });
			fixesByFile.set(fileKey, fixes);
		}

		// TS2322: null → undefined for string | undefined
		if (
			err.code === "2322" &&
			err.message.includes("null") &&
			err.message.includes("undefined")
		) {
			const m = line.match(/ref<[^>]*null[^>]*>\([^)]*null[^)]*\)/);
			if (m) {
				const fixes = fixesByFile.get(fileKey) || [];
				fixes.push({
					line: err.line,
					old: m[0],
					new: m[0]
						.replace(/null/g, "undefined")
						.replace(/string \| null/, "string | undefined"),
				});
				fixesByFile.set(fileKey, fixes);
			}
		}

		// TS2322: provider literal → as const
		if (err.code === "2322" && err.message.includes("provider")) {
			const m = line.match(/id:\s*'(\w+)'/);
			if (m) {
				const fixes = fixesByFile.get(fileKey) || [];
				fixes.push({
					line: err.line,
					old: `id: '${m[1]}'`,
					new: `id: '${m[1]}' as const`,
				});
				fixesByFile.set(fileKey, fixes);
			}
		}

		// TS2345: string | null → string | undefined
		if (
			err.code === "2345" &&
			err.message.includes("null") &&
			err.message.includes("undefined")
		) {
			const m = line.match(/ref<string \| null>\(null\)/);
			if (m) {
				const fixes = fixesByFile.get(fileKey) || [];
				fixes.push({
					line: err.line,
					old: m[0],
					new: "ref<string | undefined>(undefined)",
				});
				fixesByFile.set(fileKey, fixes);
			}
		}
	}

	for (const [file, fixes] of fixesByFile) {
		fixFile(file, fixes);
	}
	return fixesByFile.size;
}

async function main() {
	let iteration = 0;
	const maxIterations = 20;
	while (iteration < maxIterations) {
		iteration++;
		console.log(`\n--- Iteration ${iteration} ---`);
		const errors = runVueTsc();
		if (errors.length === 0) {
			console.log("✅ All type errors fixed!");
			return;
		}
		console.log(`Found ${errors.length} errors`);
		const fixed = applyFixes(errors);
		if (fixed === 0) {
			console.log("⚠ No automatic fixes applied. Remaining:");
			for (const e of errors.slice(0, 10))
				console.log(
					`  ${e.file}:${e.line} TS${e.code}: ${e.message.slice(0, 120)}`,
				);
			if (errors.length > 10)
				console.log(`  ... and ${errors.length - 10} more`);
			return;
		}
		console.log(`Applied fixes to ${fixed} file(s)`);
	}
	console.log("⚠ Max iterations reached");
}

main().catch(console.error);
