/**
 * AuthGuard Component
 *
 * Protected route wrapper that ensures users are authenticated before accessing protected pages.
 * Redirects unauthenticated users to the login flow.
 */

import { useEffect, useState } from "react";
import { auth } from "~/lib/auth";

export interface AuthGuardProps {
	/**
	 * Content to render when authenticated
	 */
	children: React.ReactNode;

	/**
	 * Optional loading component while checking auth
	 */
	fallback?: React.ReactNode;

	/**
	 * Redirect path after authentication (defaults to current path)
	 */
	redirectTo?: string;
}

/**
 * AuthGuard component - wraps protected routes
 */
export function AuthGuard({ children, fallback, redirectTo }: AuthGuardProps) {
	const [isChecking, setIsChecking] = useState(true);
	const [isAuthenticated, setIsAuthenticated] = useState(false);

	useEffect(() => {
		async function checkAuth() {
			// Check if user is authenticated
			const authenticated = auth.isAuthenticated();

			if (!authenticated) {
				// Try to refresh token if available
				const token = await auth.getAccessToken();
				if (token) {
					setIsAuthenticated(true);
				} else {
					// Not authenticated - redirect to login
					const currentPath = redirectTo || window.location.pathname;
					auth.login(currentPath);
					return;
				}
			} else {
				setIsAuthenticated(true);
			}

			setIsChecking(false);
		}

		checkAuth();
	}, [redirectTo]);

	// Show loading state
	if (isChecking) {
		return fallback || <AuthGuardFallback />;
	}

	// Show children if authenticated
	if (isAuthenticated) {
		return <>{children}</>;
	}

	// Should not reach here (user would be redirected to login)
	return null;
}

/**
 * Default loading fallback
 */
function AuthGuardFallback() {
	return (
		<div className="flex h-screen w-full items-center justify-center">
			<div className="text-center">
				<div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]" />
				<p className="mt-4 text-sm text-muted-foreground">
					Checking authentication...
				</p>
			</div>
		</div>
	);
}

/**
 * Hook to check authentication status
 */
export function useAuth() {
	const [user, setUser] = useState(auth.getUserInfo());
	const [isAuthenticated, setIsAuthenticated] = useState(
		auth.isAuthenticated(),
	);

	const refreshAuth = () => {
		setUser(auth.getUserInfo());
		setIsAuthenticated(auth.isAuthenticated());
	};

	return {
		user,
		isAuthenticated,
		login: auth.login.bind(auth),
		logout: auth.logout.bind(auth),
		refreshAuth,
	};
}

/**
 * Hook to get access token (for API calls)
 */
export function useAccessToken() {
	const [token, setToken] = useState<string | null>(null);

	useEffect(() => {
		async function fetchToken() {
			const accessToken = await auth.getAccessToken();
			setToken(accessToken);
		}
		fetchToken();
	}, []);

	return token;
}

/**
 * Hook to require authentication (redirect if not authenticated)
 */
export function useRequireAuth(redirectTo?: string) {
	useEffect(() => {
		async function checkAuth() {
			const authenticated = auth.isAuthenticated();

			if (!authenticated) {
				const token = await auth.getAccessToken();
				if (!token) {
					const currentPath = redirectTo || window.location.pathname;
					auth.login(currentPath);
				}
			}
		}

		checkAuth();
	}, [redirectTo]);

	return {
		isAuthenticated: auth.isAuthenticated(),
		user: auth.getUserInfo(),
	};
}
