/**
 * API Client Wrapper
 *
 * Provides a centralized HTTP client for communicating with the API service
 * with automatic auth header injection, error handling, and request/response transformation.
 */

import { env } from "~/env";

/**
 * Standard API error response structure
 */
export interface APIError {
	error: string;
	message: string;
	status: number;
}

/**
 * API client configuration
 */
interface APIClientConfig {
	baseURL: string;
	getAuthToken?: () => string | null;
}

/**
 * Request options for API calls
 */
export interface RequestOptions extends RequestInit {
	params?: Record<string, string | number | boolean>;
}

class APIClient {
	private baseURL: string;
	private getAuthToken?: () => string | null;

	constructor(config: APIClientConfig) {
		this.baseURL = config.baseURL;
		this.getAuthToken = config.getAuthToken;
	}

	/**
	 * Build full URL with query parameters
	 */
	private buildURL(
		path: string,
		params?: Record<string, string | number | boolean>,
	): string {
		const url = new URL(path, this.baseURL);

		if (params) {
			for (const [key, value] of Object.entries(params)) {
				url.searchParams.append(key, String(value));
			}
		}

		return url.toString();
	}

	/**
	 * Build request headers with auth token if available
	 */
	private buildHeaders(customHeaders?: HeadersInit): HeadersInit {
		const headers: HeadersInit = {
			"Content-Type": "application/json",
			...customHeaders,
		};

		// Add auth token if available
		const token = this.getAuthToken?.();
		if (token) {
			headers.Authorization = `Bearer ${token}`;
		}

		return headers;
	}

	/**
	 * Handle API response and errors
	 */
	private async handleResponse<T>(response: Response): Promise<T> {
		// Handle empty responses (204 No Content)
		if (response.status === 204) {
			return undefined as T;
		}

		const contentType = response.headers.get("content-type");
		const isJSON = contentType?.includes("application/json");

		if (!response.ok) {
			let error: APIError;

			if (isJSON) {
				error = await response.json();
			} else {
				error = {
					error: response.statusText,
					message: await response.text(),
					status: response.status,
				};
			}

			throw new APIClientError(error);
		}

		// Parse JSON response
		if (isJSON) {
			return response.json();
		}

		// Return response as text for non-JSON responses
		return response.text() as unknown as T;
	}

	/**
	 * Generic request method
	 */
	private async request<T>(
		method: string,
		path: string,
		options?: RequestOptions,
	): Promise<T> {
		const { params, headers, ...fetchOptions } = options || {};

		const url = this.buildURL(path, params);
		const requestHeaders = this.buildHeaders(headers);

		try {
			const response = await fetch(url, {
				method,
				headers: requestHeaders,
				...fetchOptions,
			});

			return this.handleResponse<T>(response);
		} catch (error) {
			// Re-throw APIClientError
			if (error instanceof APIClientError) {
				throw error;
			}

			// Wrap network errors
			throw new APIClientError({
				error: "NetworkError",
				message:
					error instanceof Error ? error.message : "Network request failed",
				status: 0,
			});
		}
	}

	/**
	 * GET request
	 */
	async get<T>(path: string, options?: RequestOptions): Promise<T> {
		return this.request<T>("GET", path, options);
	}

	/**
	 * POST request
	 */
	async post<T>(
		path: string,
		body?: unknown,
		options?: RequestOptions,
	): Promise<T> {
		return this.request<T>("POST", path, {
			...options,
			body: body ? JSON.stringify(body) : undefined,
		});
	}

	/**
	 * PATCH request
	 */
	async patch<T>(
		path: string,
		body?: unknown,
		options?: RequestOptions,
	): Promise<T> {
		return this.request<T>("PATCH", path, {
			...options,
			body: body ? JSON.stringify(body) : undefined,
		});
	}

	/**
	 * PUT request
	 */
	async put<T>(
		path: string,
		body?: unknown,
		options?: RequestOptions,
	): Promise<T> {
		return this.request<T>("PUT", path, {
			...options,
			body: body ? JSON.stringify(body) : undefined,
		});
	}

	/**
	 * DELETE request
	 */
	async delete<T>(path: string, options?: RequestOptions): Promise<T> {
		return this.request<T>("DELETE", path, options);
	}
}

/**
 * Custom error class for API client errors
 */
export class APIClientError extends Error {
	public readonly error: string;
	public readonly status: number;

	constructor(apiError: APIError) {
		super(apiError.message);
		this.name = "APIClientError";
		this.error = apiError.error;
		this.status = apiError.status;
	}

	/**
	 * Check if error is a specific HTTP status code
	 */
	is(status: number): boolean {
		return this.status === status;
	}

	/**
	 * Check if error is unauthorized (401)
	 */
	isUnauthorized(): boolean {
		return this.status === 401;
	}

	/**
	 * Check if error is forbidden (403)
	 */
	isForbidden(): boolean {
		return this.status === 403;
	}

	/**
	 * Check if error is not found (404)
	 */
	isNotFound(): boolean {
		return this.status === 404;
	}

	/**
	 * Check if error is rate limited (429)
	 */
	isRateLimited(): boolean {
		return this.status === 429;
	}

	/**
	 * Check if error is a server error (5xx)
	 */
	isServerError(): boolean {
		return this.status >= 500 && this.status < 600;
	}
}

/**
 * Default API client instance
 * Integrates with auth manager for automatic token injection
 */
export const apiClient = new APIClient({
	baseURL: env.VITE_API_URL,
	getAuthToken: () => {
		// Import auth manager dynamically to avoid circular dependencies
		if (typeof window !== "undefined") {
			const authState = localStorage.getItem("refract_auth_state");
			if (authState) {
				try {
					const parsed = JSON.parse(authState);
					return parsed.access_token || null;
				} catch {
					return null;
				}
			}
		}
		return null;
	},
});

/**
 * Create API client with custom config
 */
export function createAPIClient(config: Partial<APIClientConfig>): APIClient {
	return new APIClient({
		baseURL: config.baseURL || env.VITE_API_URL,
		getAuthToken: config.getAuthToken,
	});
}
