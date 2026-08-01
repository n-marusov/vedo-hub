// @ctx: Runtime configuration reader — window.__VEDO_CONFIG__ is generated
// by docker-entrypoint.sh at container startup (served from /config.js).
// Values are non-sensitive public config; PUBLIC_DOMAIN comes from the
// VEDO_PUBLIC_DOMAIN environment variable.

const DEFAULT_PUBLIC_DOMAIN = "vedo-core.local";

/**
 * Public domain/host used as the immutable prefix when building group and
 * project URLs. Resolved at runtime so the same bundle works across
 * environments (dev, test, staging, production) without a rebuild.
 */
export function getPublicDomain(): string {
	return window.__VEDO_CONFIG__?.PUBLIC_DOMAIN?.trim() || DEFAULT_PUBLIC_DOMAIN;
}
