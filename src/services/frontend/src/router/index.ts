import { isAuthenticated, isTokenValid } from "@/auth/session";
import { logAuthRedirect, logNavigation } from "@/utils/structured-logger";
// @ctx: vue-router configuration for all application screens per GUI-OW-001
// @hlv:artifact router-index implements GUI-OW-001
import { createRouter, createWebHistory } from "vue-router";
import type {
	NavigationGuardNext,
	RouteLocationNormalized,
	RouteRecordRaw,
} from "vue-router";

// @ctx: route definitions per GUI-OW-001 Input route enum
const routes: RouteRecordRaw[] = [
	// @hlv GUI_ROUTE_NOT_FOUND
	{
		path: "/",
		name: "root",
		redirect: "/dashboard/home",
	},
	{
		path: "/login",
		name: "login",
		component: () => import("@/pages/LoginPage.vue"),
		meta: { requiresAuth: false, title: "Login" },
	},
	{
		path: "/auth/callback",
		name: "auth-callback",
		component: () => import("@/pages/AuthCallbackPage.vue"),
		meta: { requiresAuth: false, title: "Authenticating" },
	},
	{
		path: "/dashboard",
		redirect: { name: "dashboard-home" },
	},
	{
		path: "/dashboard/home",
		name: "dashboard-home",
		component: () => import("@/pages/DashboardPage.vue"),
		meta: { requiresAuth: true, title: "Dashboard" },
	},
	{
		path: "/ontology/:id/workspace",
		name: "ontology-workspace",
		component: () => import("@/pages/OntologyWorkspace.vue"),
		meta: { requiresAuth: true, title: "Ontology Workspace" },
	},
	{
		path: "/ontology/:id/query",
		name: "ontology-query",
		alias: "/ontology/:id/sparql",
		component: () => import("@/pages/SPARQLPage.vue"),
		meta: { requiresAuth: true, title: "SPARQL Query Builder" },
	},
	{
		path: "/dashboard/groups",
		name: "groups",
		component: () => import("@/pages/GroupsPage.vue"),
		meta: { requiresAuth: true, title: "Groups" },
	},
	{
		path: "/dashboard/projects",
		name: "projects",
		component: () => import("@/pages/ProjectsPage.vue"),
		meta: { requiresAuth: true, title: "Projects" },
	},
	{
		path: "/dashboard/merge_requests",
		name: "merge-requests",
		component: () => import("@/pages/MergeRequestsPage.vue"),
		meta: { requiresAuth: true, title: "Merge Requests" },
	},
	{
		path: "/commits",
		name: "commits",
		component: () => import("@/pages/RecentCommitsPage.vue"),
		meta: { requiresAuth: true, title: "Commits" },
	},
	{
		path: "/recent-commits",
		name: "recent-commits",
		component: () => import("@/pages/RecentCommitsPage.vue"),
		meta: { requiresAuth: true, title: "Commits" },
	},
	{
		path: "/comments",
		name: "comments",
		component: () => import("@/pages/CommentsPage.vue"),
		meta: { requiresAuth: true, title: "Comments" },
	},
	{
		path: "/dashboard/deployments",
		name: "deployments",
		component: () => import("@/pages/DeploymentsPage.vue"),
		meta: { requiresAuth: true, title: "Deployments" },
	},
	{
		path: "/metrics",
		name: "metrics",
		component: () => import("@/pages/MetricsPage.vue"),
		meta: { requiresAuth: true, title: "Metrics" },
	},
	{
		path: "/ontology/:id/members",
		name: "ontology-members",
		component: () => import("@/pages/MembersPage.vue"),
		meta: { requiresAuth: true, title: "Members" },
	},
	{
		path: "/ontology/:id/validation",
		name: "ontology-validation",
		component: () => import("@/pages/ValidationPage.vue"),
		meta: { requiresAuth: true, title: "Validation" },
	},
	{
		path: "/ontology/:id/shacl",
		name: "ontology-shacl",
		component: () => import("@/pages/ShaclPage.vue"),
		meta: { requiresAuth: true, title: "SHACL Rule Builder" },
	},
	{
		path: "/ontology/:id/versioning/:view(commits|branches|compare|tags|graph|merge_requests)",
		name: "ontology-versioning",
		component: () => import("@/pages/VersioningPage.vue"),
		meta: { requiresAuth: true, title: "Versioning" },
	},
	{
		path: "/public/:id",
		name: "public-ontology",
		component: () => import("@/pages/PublicOntologyPage.vue"),
		meta: { requiresAuth: false, title: "Public Ontology" },
	},
	// @hlv GUI_ROUTE_NOT_FOUND
	{
		path: "/:pathMatch(.*)*",
		name: "not-found",
		component: () => import("@/pages/NotFoundPage.vue"),
		meta: { requiresAuth: false, title: "Not Found" },
	},
];

const router = createRouter({
	history: createWebHistory(),
	routes,
});

// @ctx: auth guard per GUI-OW-001 Invariant 1 — all protected routes must have auth guard
// @hlv GUI_AUTH_REQUIRED
router.beforeEach(
	(
		to: RouteLocationNormalized,
		_from: RouteLocationNormalized,
		next: NavigationGuardNext,
	) => {
		logNavigation({ route: to.name as string, path: to.path });

		if (to.meta.requiresAuth) {
			if (!isAuthenticated()) {
				logAuthRedirect({ reason: "unauthenticated", target: to.path });
				// @hlv GUI_AUTH_REQUIRED
				next({ name: "login", query: { redirect: to.fullPath } });
				return;
			}
			if (!isTokenValid()) {
				logAuthRedirect({ reason: "token_expired", target: to.path });
				// @hlv GUI_AUTH_REQUIRED
				next({ name: "login", query: { redirect: to.fullPath } });
				return;
			}
		}

		// @ctx: redirect authenticated users away from login
		if (to.name === "login" && isAuthenticated()) {
			next({ name: "dashboard" });
			return;
		}

		next();
	},
);

export default router;
