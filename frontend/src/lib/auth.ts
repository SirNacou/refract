/**
 * Zitadel OIDC Authentication Integration
 *
 * Implements OAuth2/OIDC authentication flow with PKCE for Zitadel identity provider.
 * Handles login, logout, token refresh, and token storage.
 */

import { env } from "~/env";

/**
 * OIDC token response structure
 */
export interface TokenResponse {
	access_token: string;
	token_type: string;
	expires_in: number;
	refresh_token?: string;
	id_token?: string;
	scope?: string;
}

/**
 * Decoded JWT payload (subset of claims)
 */
export interface JWTPayload {
	sub: string; // Subject (user ID from Zitadel)
	email?: string;
	email_verified?: boolean;
	name?: string;
	given_name?: string;
	family_name?: string;
	preferred_username?: string;
	iat: number; // Issued at
	exp: number; // Expiration
	iss: string; // Issuer
	aud: string | string[]; // Audience
}

/**
 * Auth state stored in localStorage
 */
interface AuthState {
	access_token: string;
	refresh_token?: string;
	id_token?: string;
	expires_at: number; // Unix timestamp
}

/**
 * PKCE challenge and verifier
 */
interface PKCEPair {
	verifier: string;
	challenge: string;
}

/**
 * Storage keys for auth data
 */
const STORAGE_KEYS = {
	AUTH_STATE: "refract_auth_state",
	PKCE_VERIFIER: "refract_pkce_verifier",
	REDIRECT_PATH: "refract_redirect_path",
} as const;

/**
 * OIDC configuration
 */
const OIDC_CONFIG = {
	authority: env.VITE_ZITADEL_AUTHORITY,
	clientId: env.VITE_ZITADEL_CLIENT_ID,
	redirectUri:
		typeof window !== "undefined"
			? `${window.location.origin}/auth/callback`
			: "",
	postLogoutRedirectUri:
		typeof window !== "undefined" ? window.location.origin : "",
	scope: "openid profile email offline_access", // offline_access for refresh token
	responseType: "code",
} as const;

/**
 * Generate random string for PKCE verifier
 */
function generateRandomString(length: number): string {
	const charset =
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~";
	const randomValues = new Uint8Array(length);
	crypto.getRandomValues(randomValues);
	return Array.from(randomValues)
		.map((v) => charset[v % charset.length])
		.join("");
}

/**
 * Generate SHA-256 hash
 */
async function sha256(plain: string): Promise<ArrayBuffer> {
	const encoder = new TextEncoder();
	const data = encoder.encode(plain);
	return crypto.subtle.digest("SHA-256", data);
}

/**
 * Base64 URL encode
 */
function base64URLEncode(buffer: ArrayBuffer): string {
	const bytes = new Uint8Array(buffer);
	let binary = "";
	for (let i = 0; i < bytes.byteLength; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

/**
 * Generate PKCE code verifier and challenge
 */
async function generatePKCE(): Promise<PKCEPair> {
	const verifier = generateRandomString(128);
	const challengeBuffer = await sha256(verifier);
	const challenge = base64URLEncode(challengeBuffer);

	return { verifier, challenge };
}

/**
 * Decode JWT without verification (verification happens on backend)
 */
function decodeJWT(token: string): JWTPayload | null {
	try {
		const parts = token.split(".");
		if (parts.length !== 3) return null;

		const payload = parts[1];
		const decoded = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
		return JSON.parse(decoded);
	} catch {
		return null;
	}
}

/**
 * Storage helpers (only run in browser)
 */
const storage = {
	get<T>(key: string): T | null {
		if (typeof window === "undefined") return null;
		try {
			const item = localStorage.getItem(key);
			return item ? JSON.parse(item) : null;
		} catch {
			return null;
		}
	},

	set(key: string, value: unknown): void {
		if (typeof window === "undefined") return;
		try {
			localStorage.setItem(key, JSON.stringify(value));
		} catch (error) {
			console.error("Failed to save to localStorage:", error);
		}
	},

	remove(key: string): void {
		if (typeof window === "undefined") return;
		localStorage.removeItem(key);
	},
};

/**
 * Auth manager class
 */
class AuthManager {
	/**
	 * Start login flow (redirect to Zitadel)
	 */
	async login(redirectPath?: string): Promise<void> {
		// Save redirect path for post-login navigation
		if (redirectPath) {
			storage.set(STORAGE_KEYS.REDIRECT_PATH, redirectPath);
		}

		// Generate PKCE pair
		const pkce = await generatePKCE();
		storage.set(STORAGE_KEYS.PKCE_VERIFIER, pkce.verifier);

		// Build authorization URL
		const params = new URLSearchParams({
			client_id: OIDC_CONFIG.clientId,
			redirect_uri: OIDC_CONFIG.redirectUri,
			response_type: OIDC_CONFIG.responseType,
			scope: OIDC_CONFIG.scope,
			code_challenge: pkce.challenge,
			code_challenge_method: "S256",
			state: generateRandomString(32), // CSRF protection
		});

		const authUrl = `${OIDC_CONFIG.authority}/oauth/v2/authorize?${params.toString()}`;
		window.location.href = authUrl;
	}

	/**
	 * Handle callback from Zitadel (exchange code for tokens)
	 */
	async handleCallback(code: string): Promise<void> {
		// Retrieve PKCE verifier
		const verifier = storage.get<string>(STORAGE_KEYS.PKCE_VERIFIER);
		if (!verifier) {
			throw new Error("PKCE verifier not found. Please try logging in again.");
		}

		// Exchange authorization code for tokens
		const tokenResponse = await this.exchangeCode(code, verifier);

		// Save auth state
		this.saveAuthState(tokenResponse);

		// Clean up PKCE verifier
		storage.remove(STORAGE_KEYS.PKCE_VERIFIER);
	}

	/**
	 * Exchange authorization code for tokens
	 */
	private async exchangeCode(
		code: string,
		verifier: string,
	): Promise<TokenResponse> {
		const params = new URLSearchParams({
			grant_type: "authorization_code",
			client_id: OIDC_CONFIG.clientId,
			code,
			redirect_uri: OIDC_CONFIG.redirectUri,
			code_verifier: verifier,
		});

		const response = await fetch(`${OIDC_CONFIG.authority}/oauth/v2/token`, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
			},
			body: params.toString(),
		});

		if (!response.ok) {
			const error = await response.text();
			throw new Error(`Token exchange failed: ${error}`);
		}

		return response.json();
	}

	/**
	 * Save auth state to storage
	 */
	private saveAuthState(tokenResponse: TokenResponse): void {
		const expiresAt = Date.now() + tokenResponse.expires_in * 1000;

		const authState: AuthState = {
			access_token: tokenResponse.access_token,
			refresh_token: tokenResponse.refresh_token,
			id_token: tokenResponse.id_token,
			expires_at: expiresAt,
		};

		storage.set(STORAGE_KEYS.AUTH_STATE, authState);
	}

	/**
	 * Get current access token (refresh if expired)
	 */
	async getAccessToken(): Promise<string | null> {
		const authState = storage.get<AuthState>(STORAGE_KEYS.AUTH_STATE);
		if (!authState) return null;

		// Check if token is expired (with 5-minute buffer)
		const expiresIn = authState.expires_at - Date.now();
		if (expiresIn > 5 * 60 * 1000) {
			return authState.access_token;
		}

		// Try to refresh token
		if (authState.refresh_token) {
			try {
				await this.refreshToken(authState.refresh_token);
				const newAuthState = storage.get<AuthState>(STORAGE_KEYS.AUTH_STATE);
				return newAuthState?.access_token || null;
			} catch (error) {
				console.error("Token refresh failed:", error);
				this.logout(); // Clear invalid session
				return null;
			}
		}

		// Token expired and no refresh token
		return null;
	}

	/**
	 * Refresh access token using refresh token
	 */
	private async refreshToken(refreshToken: string): Promise<void> {
		const params = new URLSearchParams({
			grant_type: "refresh_token",
			client_id: OIDC_CONFIG.clientId,
			refresh_token: refreshToken,
		});

		const response = await fetch(`${OIDC_CONFIG.authority}/oauth/v2/token`, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
			},
			body: params.toString(),
		});

		if (!response.ok) {
			throw new Error("Token refresh failed");
		}

		const tokenResponse: TokenResponse = await response.json();
		this.saveAuthState(tokenResponse);
	}

	/**
	 * Get current user info from ID token
	 */
	getUserInfo(): JWTPayload | null {
		const authState = storage.get<AuthState>(STORAGE_KEYS.AUTH_STATE);
		if (!authState?.id_token) return null;

		return decodeJWT(authState.id_token);
	}

	/**
	 * Check if user is authenticated
	 */
	isAuthenticated(): boolean {
		const authState = storage.get<AuthState>(STORAGE_KEYS.AUTH_STATE);
		if (!authState) return false;

		// Check if token is still valid
		return authState.expires_at > Date.now();
	}

	/**
	 * Logout user (clear local state and redirect to Zitadel logout)
	 */
	async logout(): Promise<void> {
		const authState = storage.get<AuthState>(STORAGE_KEYS.AUTH_STATE);

		// Clear local storage
		storage.remove(STORAGE_KEYS.AUTH_STATE);
		storage.remove(STORAGE_KEYS.REDIRECT_PATH);

		// Redirect to Zitadel logout endpoint
		if (authState?.id_token) {
			const params = new URLSearchParams({
				id_token_hint: authState.id_token,
				post_logout_redirect_uri: OIDC_CONFIG.postLogoutRedirectUri,
			});

			const logoutUrl = `${OIDC_CONFIG.authority}/oidc/v1/end_session?${params.toString()}`;
			window.location.href = logoutUrl;
		} else {
			// No ID token, just redirect to home
			window.location.href = "/";
		}
	}

	/**
	 * Get saved redirect path after login
	 */
	getRedirectPath(): string | null {
		return storage.get<string>(STORAGE_KEYS.REDIRECT_PATH);
	}

	/**
	 * Clear saved redirect path
	 */
	clearRedirectPath(): void {
		storage.remove(STORAGE_KEYS.REDIRECT_PATH);
	}
}

/**
 * Export singleton instance
 */
export const auth = new AuthManager();

/**
 * Export for testing/custom instances
 */
export { AuthManager };
