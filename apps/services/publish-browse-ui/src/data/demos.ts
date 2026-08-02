// Demo ontology showcase data for the public landing page.
//
// F11.1 — the MVP demo chain needs clickable demo ontologies without auth.
// The publish-browse-ui nginx serves static files only (no /api proxy), so the
// showcase is self-contained: class structures are derived from the VEDO Demos
// seed fixtures (deploy/seeds/vedo-demos/*.json) and bundled here. When the
// public-browse-api gains real snapshots, this module can be replaced by a
// fetch against GET /api/v1/ontologies without changing the UI.

export interface DemoOntology {
	id: string;
	name: string;
	description: string;
	classCount: number;
	classTree: DemoClassNode[];
}

export interface DemoClassNode {
	id: string;
	label: string;
	children: DemoClassNode[];
}

// Class trees mirroring the VEDO Demos seed fixtures.
const productTree: DemoClassNode[] = [
	{ id: "Product", label: "Product", children: [] },
	{ id: "Category", label: "Category", children: [] },
	{ id: "Manufacturer", label: "Manufacturer", children: [] },
	{ id: "Review", label: "Review", children: [] },
];

const organizationTree: DemoClassNode[] = [
	{ id: "Organization", label: "Organization", children: [] },
	{ id: "Department", label: "Department", children: [] },
	{ id: "Employee", label: "Employee", children: [] },
	{ id: "Role", label: "Role", children: [] },
];

const processTree: DemoClassNode[] = [
	{ id: "Process", label: "Process", children: [] },
	{ id: "Activity", label: "Activity", children: [] },
	{ id: "Input", label: "Input", children: [] },
	{ id: "Output", label: "Output", children: [] },
];

const glossaryTree: DemoClassNode[] = [
	{ id: "Term", label: "Term", children: [] },
	{ id: "Definition", label: "Definition", children: [] },
	{ id: "Synonym", label: "Synonym", children: [] },
	{ id: "Acronym", label: "Acronym", children: [] },
];

const eventTree: DemoClassNode[] = [
	{ id: "Event", label: "Event", children: [] },
	{ id: "Participant", label: "Participant", children: [] },
	{ id: "Venue", label: "Venue", children: [] },
	{ id: "Schedule", label: "Schedule", children: [] },
];

export const VEDO_DEMOS: DemoOntology[] = [
	{
		id: "demo-product",
		name: "Продукт",
		description:
			"Продуктовая онтология: товары, категории, производители и отзывы.",
		classCount: 4,
		classTree: productTree,
	},
	{
		id: "demo-organization",
		name: "Организация",
		description: "Организационная структура: отделы, сотрудники, роли.",
		classCount: 4,
		classTree: organizationTree,
	},
	{
		id: "demo-process",
		name: "Процесс",
		description: "Онтология процессов: активности, входы и выходы.",
		classCount: 4,
		classTree: processTree,
	},
	{
		id: "demo-glossary",
		name: "Глоссарий",
		description: "Термины, определения, синонимы и аббревиатуры.",
		classCount: 4,
		classTree: glossaryTree,
	},
	{
		id: "demo-event",
		name: "Событие",
		description: "События, участники, площадки и расписания.",
		classCount: 4,
		classTree: eventTree,
	},
];

export function findDemo(id: string): DemoOntology | undefined {
	return VEDO_DEMOS.find((d) => d.id === id);
}
