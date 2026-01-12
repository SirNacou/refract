# Feature Specification: Authenticated URL Shortener Platform

**Feature Branch**: `016-authenticated-url-shortener`  
**Created**: 2026-01-09  
**Status**: Draft  
**Input**: User description: "Build an url shortener that only authenticated user can create urls. And they can see analysis of clicks. Make this project follow all best practices of other url shortener. No billing feature. Distributed system. Only frontend is tanstack start. Optionally user can use apikey to create shortlink from code"

## Clarifications

### Session 2026-01-09

- Q: Should authentication be built in-house or use an identity provider (Zitadel deployed)? → A: Use Zitadel as IdP via OIDC/OAuth2
- Q: What database strategy for analytics data (high volume, time-series queries)? → A: Time-series database (TimescaleDB or ClickHouse)
- Q: What database ID type should be used to generate short codes? → A: Snowflake IDs (64-bit distributed) with Base62 encoding
- Q: What caching architecture for <50ms redirect latency? → A: Multi-tier caching (in-memory L1 + Redis/Valkey L2)
- Q: How should API keys be stored and validated securely? → A: Hash API keys using BLAKE2 or SHA256
- Q: What TanStack Start project structure should be used? → A: Use official TanStack Start scaffold structure (npx create-tanstack@latest)
- Q: Should Zitadel be included in docker-compose.yml or reference external deployment? → A: Reference existing external Zitadel via environment variables

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Share Short URLs (Priority: P1)

A logged-in user wants to convert a long URL into a short, shareable link that they can distribute via social media, email, or messaging apps.

**Why this priority**: This is the core value proposition - without URL shortening, there is no product. This must work before any other features matter.

**Independent Test**: User can log in, paste a long URL, receive a short URL, and successfully redirect to the original destination when visiting the short link.

**Acceptance Scenarios**:

1. **Given** a logged-in user on the dashboard, **When** they paste "https://example.com/very/long/path/to/resource?param1=value1&param2=value2" and click "Shorten", **Then** they receive a short URL like "https://short.link/abc123" that redirects to the original URL
2. **Given** a user receives a short URL, **When** anyone (authenticated or not) visits that short URL in a browser, **Then** they are immediately redirected to the original long URL
3. **Given** a user creates a short URL, **When** they set a custom alias (e.g., "my-blog"), **Then** the short URL uses their custom alias if available (e.g., "https://short.link/my-blog")
4. **Given** a user tries to create a short URL with an already-taken custom alias, **When** they submit the form, **Then** they receive a clear error message and can choose a different alias
5. **Given** an unauthenticated visitor, **When** they try to access the URL creation page, **Then** they are redirected to the login page

---

### User Story 2 - View Click Analytics (Priority: P2)

A user wants to understand how their short URLs are performing by viewing detailed click analytics including click counts, geographic locations, referrer sources, and time-based trends.

**Why this priority**: Analytics provide the "why" for using this service over competitors. Users need data to understand their audience and optimize their content distribution.

**Independent Test**: User can create a short URL, generate clicks from different sources/locations, then view a dashboard showing accurate click statistics with filters and visualizations.

**Acceptance Scenarios**:

1. **Given** a logged-in user with existing short URLs, **When** they navigate to their dashboard, **Then** they see a list of all their URLs with total click counts for each
2. **Given** a user viewing analytics for a specific short URL, **When** they access the detailed analytics page, **Then** they see click count over time (hourly/daily/weekly views), geographic distribution (country/city), referrer sources (direct, social media, email), device types (desktop/mobile/tablet), and browser information
3. **Given** a short URL receives clicks, **When** the user refreshes their analytics dashboard, **Then** the click data updates within 5 seconds (near real-time)
4. **Given** a user viewing analytics, **When** they select a date range filter (e.g., "Last 7 days"), **Then** all metrics update to reflect only clicks within that timeframe
5. **Given** a user has no URLs created yet, **When** they access their dashboard, **Then** they see an empty state with a call-to-action to create their first short URL

---

### User Story 3 - Manage Short URLs (Priority: P3)

A user wants to organize, edit, and delete their short URLs to maintain control over their link portfolio.

**Why this priority**: As users create more links, they need management capabilities. This enables long-term use but isn't required for initial value delivery.

**Independent Test**: User can view their URL list, edit URL properties (title, custom alias), deactivate URLs to stop redirects, and delete URLs they no longer need.

**Acceptance Scenarios**:

1. **Given** a logged-in user with multiple short URLs, **When** they view their dashboard, **Then** they can search/filter URLs by title, destination, or creation date
2. **Given** a user viewing a specific short URL, **When** they click "Edit" and update the title or destination URL, **Then** changes are saved and reflected immediately
3. **Given** a user with a live short URL, **When** they deactivate/disable the URL, **Then** visitors to that short link see a "Link Disabled" page instead of redirecting
4. **Given** a user wants to remove a short URL permanently, **When** they delete it, **Then** the URL is removed from their dashboard and shows a "404 Not Found" page to visitors
5. **Given** a user views their URL list, **When** they sort by click count, creation date, or last modified, **Then** the list reorders accordingly

---

### User Story 4 - Programmatic URL Creation via API Key (Priority: P4)

A developer or power user wants to create short URLs programmatically from their applications, scripts, or automation tools using an API key for authentication.

**Why this priority**: This enables advanced use cases (bulk shortening, integration with content management systems, automation workflows) but the core web UI delivers the primary value.

**Independent Test**: User can generate an API key from their dashboard, use it to create short URLs via HTTP requests, and see those URLs appear in their dashboard alongside manually-created ones.

**Acceptance Scenarios**:

1. **Given** a logged-in user, **When** they navigate to "API Keys" settings, **Then** they can generate a new API key with a descriptive name (e.g., "Blog automation")
2. **Given** a user has an active API key, **When** they send a POST request with the API key and long URL, **Then** they receive a JSON response containing the short URL and metadata
3. **Given** a user with multiple API keys, **When** they view their API keys list, **Then** they see key names, creation dates, last used timestamps, and usage counts
4. **Given** a user suspects an API key is compromised, **When** they revoke/delete the key, **Then** all subsequent requests with that key are rejected with a 401 Unauthorized error
5. **Given** an API key is used to create a short URL, **When** the user views their dashboard, **Then** they can see which URLs were created via API (with an indicator/badge)

---

### User Story 5 - User Authentication and Account Management (Priority: P1)

A new visitor wants to create an account and log in to access the URL shortening service via Zitadel identity provider, ensuring their links are private and associated with their identity.

**Why this priority**: Authentication is a prerequisite for P1 (creating URLs requires login). This is foundational infrastructure that enables all other stories.

**Independent Test**: Visitor can sign up via Zitadel (email/password or OAuth providers), log in, access protected pages, log out, and reset password through Zitadel flows.

**Acceptance Scenarios**:

1. **Given** a new visitor on the homepage, **When** they click "Sign Up" and are redirected to Zitadel, **Then** they can register via email/password or OAuth providers (Google, GitHub) and return to the application authenticated
2. **Given** a user with a Zitadel account, **When** they click "Log In" and authenticate via Zitadel, **Then** they receive a JWT token and are redirected to their dashboard
3. **Given** a logged-in user, **When** they click "Log Out", **Then** their Zitadel session ends and they are redirected to the homepage (no longer able to access protected pages)
4. **Given** a user forgot their password, **When** they request a password reset, **Then** Zitadel handles the reset flow via email
5. **Given** a user tries to sign up with an already-registered email, **When** they submit the form to Zitadel, **Then** Zitadel shows an error message: "Email already registered"

---

### Edge Cases

- **What happens when a user creates a short URL for a malicious/phishing destination?** System must validate URLs against safe browsing APIs and block known malicious domains
- **What happens when a short URL receives extremely high traffic (DDoS or viral content)?** System must use distributed caching and rate limiting to maintain performance and prevent abuse
- **What happens when a user tries to create thousands of short URLs in rapid succession?** System must enforce rate limits (e.g., 100 URLs per hour for free users) to prevent abuse
- **What happens when the same long URL is shortened by multiple users?** Each user gets their own unique short code and independent analytics (no data sharing between users)
- **What happens when a custom alias contains special characters, spaces, or reserved words?** System must validate aliases (alphanumeric + hyphens only, no reserved words like "admin", "api", "dashboard")
- **What happens when a short URL expires or is deactivated?** Visitor sees a branded "Link Unavailable" page with explanation (expired, disabled by owner, or deleted)
- **What happens when a user deletes their account?** All their short URLs are deactivated, analytics data is anonymized/retained for aggregate statistics, and personally identifiable information is deleted
- **What happens when analytics tracking is blocked by ad blockers or privacy tools?** System gracefully degrades, logging what data is available (at minimum, server-side IP-based geolocation)
- **What happens when two users try to claim the same custom alias simultaneously?** System uses database-level unique constraints to ensure first request wins, second receives error
- **What happens when the original destination URL becomes unavailable (404/500)?** System still redirects (it's a mirror/proxy service, not a validator), but may show a banner: "Warning: This destination may be unavailable"

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication & Authorization

- **FR-001**: System MUST delegate authentication to Zitadel identity provider via OIDC/OAuth2
- **FR-002**: System MUST validate JWT tokens issued by Zitadel for all authenticated requests
- **FR-003**: System MUST extract user identity (sub, email) from Zitadel JWT claims
- **FR-004**: System MUST handle Zitadel authentication flows (authorization code flow for web, PKCE for SPAs)
- **FR-005**: System MUST allow users to log out by invalidating Zitadel session and clearing local tokens
- **FR-006**: System MUST allow unauthenticated visitors to access short URLs (redirects work for everyone)
- **FR-007**: Zitadel MUST be configured to support email/password and OAuth providers (Google, GitHub) for sign-up/login

#### URL Shortening

- **FR-008**: System MUST generate Snowflake IDs (64-bit distributed) for each URL and encode them to Base62 for short codes (8-11 characters, URL-safe alphabet)
- **FR-009**: System MUST allow users to specify custom aliases (if available) instead of auto-generated codes
- **FR-010**: System MUST validate custom aliases (3-50 characters, alphanumeric and hyphens only, no reserved words)
- **FR-011**: System MUST enforce unique short codes globally (no collisions between users)
- **FR-012**: System MUST validate destination URLs against safe browsing/phishing databases before creation
- **FR-013**: System MUST support optional URL expiration dates (user-specified time-to-live)
- **FR-014**: System MUST allow users to set optional metadata (title, notes) for organizational purposes

#### URL Redirection

- **FR-015**: System MUST redirect visitors from short URL to destination URL with HTTP 301 (permanent) or 302 (temporary) status codes
- **FR-016**: System MUST perform redirects in under 50 milliseconds at 95th percentile (p95 latency)
- **FR-017**: System MUST implement multi-tier caching: in-memory L1 cache for hot URLs (<1ms latency) and Redis/Valkey L2 cache for distributed access (2-5ms latency)
- **FR-018**: System MUST handle expired URLs by showing an "Expired Link" page instead of redirecting
- **FR-019**: System MUST handle deactivated URLs by showing a "Link Disabled" page

#### Analytics & Tracking

- **FR-020**: System MUST capture click events including timestamp, referrer, user agent, IP address (anonymized for privacy)
- **FR-021**: System MUST derive geographic location (country, region, city) from IP addresses
- **FR-022**: System MUST extract device type (desktop, mobile, tablet) and browser information from user agents
- **FR-023**: System MUST aggregate click data into time-series metrics (hourly, daily, weekly, monthly views)
- **FR-024**: System MUST display near real-time analytics (event-to-store processing within 5 seconds, API query response <100ms p95)
- **FR-025**: System MUST provide filtering/date range selection for analytics views
- **FR-026**: System MUST show top referrers (domains that linked to the short URL)
- **FR-027**: System MUST track unique visitors vs total clicks (deduplicate by IP + user agent within 24-hour window)

#### URL Management

- **FR-028**: System MUST allow users to view a list of all their short URLs with pagination
- **FR-029**: System MUST allow users to search/filter their URLs by title, destination, or date
- **FR-030**: System MUST allow users to edit URL metadata (title, notes, destination URL)
- **FR-031**: System MUST allow users to deactivate/disable URLs without deleting them
- **FR-032**: System MUST allow users to permanently delete URLs and their analytics data
- **FR-033**: System MUST show URL status indicators (active, expired, disabled, deleted)

#### API Access

- **FR-034**: System MUST allow users to generate API keys from their account settings, storing them hashed (BLAKE2 or SHA256) with key prefix for identification
- **FR-035**: System MUST support creating short URLs via RESTful API using API key authentication (hashed key validation)
- **FR-036**: System MUST allow users to name, view, and revoke API keys
- **FR-037**: System MUST track API key usage (last used timestamp, request counts)
- **FR-038**: System MUST enforce the same rate limits and validation rules for API-created URLs
- **FR-039**: System MUST return structured JSON responses for API requests with appropriate HTTP status codes

#### Security & Abuse Prevention

- **FR-040**: System MUST enforce rate limiting on URL creation using token bucket algorithm (100 URLs per hour per user, Redis-backed counters, return HTTP 429 with Retry-After header)
- **FR-041**: System MUST enforce rate limiting on API requests using token bucket algorithm (1000 requests per hour per API key, Redis-backed counters, return HTTP 429 with Retry-After header)
- **FR-042**: System MUST block creation of short URLs pointing to known malicious/phishing domains
- **FR-043**: System MUST use HTTPS for all connections (frontend and API)
- **FR-044**: System MUST NOT store passwords (authentication delegated to Zitadel identity provider)
- **FR-045**: System MUST implement CSRF protection for web forms
- **FR-046**: System MUST validate and sanitize all user inputs to prevent injection attacks

#### System Architecture

- **FR-047**: System MUST be designed as a distributed system with multiple services (API, redirector, analytics processor)
- **FR-048**: System MUST use caching layers to serve redirects at scale
- **FR-049**: System MUST use asynchronous event processing for analytics data (decouple click tracking from redirects)
- **FR-050**: System MUST support horizontal scaling of services independently
- **FR-051**: System MUST use time-series optimized database (TimescaleDB or ClickHouse) for click events and analytics aggregations

### Key Entities

- **User**: Represents an authenticated account holder managed by Zitadel. Attributes: user ID (Zitadel subject ID), email, account creation date, Zitadel metadata sync timestamp. Relationships: owns multiple URLs and API keys. Note: Password management, email verification, and password reset handled by Zitadel.

- **Short URL**: Represents a shortened link. Attributes: Snowflake ID (64-bit, primary key), short code (Base62-encoded Snowflake ID), custom alias (optional, unique), destination URL, creation timestamp, expiration date (optional), status (active/expired/disabled/deleted), title, notes, creator user ID (Zitadel subject). Relationships: belongs to one user, has multiple click events.

- **Click Event**: Represents a single redirect/click stored in time-series database. Attributes: event ID, short URL ID, timestamp, referrer URL, user agent string, IP address (anonymized), geographic data (country, region, city), device type, browser type. Relationships: belongs to one short URL.

- **API Key**: Represents programmatic access credentials. Attributes: key ID, key value (hashed via BLAKE2/SHA256), key prefix (first 8 chars for identification), key name/description, creation date, last used timestamp, usage count, status (active/revoked), owner user ID. Relationships: belongs to one user, used to create multiple URLs.

- **Analytics Summary**: Represents aggregated metrics for a short URL (computed from time-series data). Attributes: short URL ID, time period (hour/day/week/month), total clicks, unique visitors, top referrers (list), geographic distribution (map), device breakdown (percentages), browser breakdown (percentages). Relationships: belongs to one short URL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a short URL and receive a working redirect in under 2 seconds from form submission
- **SC-002**: System handles 10,000 concurrent redirects without degradation (sub-50ms p95 latency maintained)
- **SC-003**: Click analytics data appears in user dashboard within 5 seconds of redirect event (measures event-to-store processing latency, not including UI query time)
- **SC-004**: 95% of users successfully create their first short URL within 3 minutes of account activation
- **SC-005**: System achieves 99.9% uptime for redirect service (core functionality)
- **SC-006**: Short URL redirects work correctly for 100% of valid destination URLs
- **SC-007**: API key creation and first API request succeed within 2 minutes for 90% of users
- **SC-008**: Zero data breaches or unauthorized access to user analytics data
- **SC-009**: System blocks 100% of short URLs pointing to known malicious domains (based on safe browsing APIs)
- **SC-010**: Users can view analytics for URLs with 1 million+ clicks without performance degradation (page loads in under 2 seconds)

## Assumptions

1. **Identity Provider**: Assumes Zitadel identity provider is already deployed externally (separate Docker containers) and accessible via environment variables. Zitadel must be configured with email/password and OAuth2 providers (Google, GitHub).
2. **Domain Name**: Assumes a short, memorable domain is available for the service (e.g., "short.link", "go.to")
3. **Safe Browsing API**: Assumes integration with Google Safe Browsing or similar service for malicious URL detection
4. **Geographic Data**: Assumes access to IP geolocation database/service (e.g., MaxMind GeoIP, IP2Location)
5. **User Capacity**: Assumes initial capacity planning for 10,000 active users creating average 50 URLs each
6. **Data Retention**: Analytics data retained indefinitely unless user deletes URL or closes account
7. **Free Service**: No payment processing, billing, or subscription tiers (all users have same feature access)
8. **Language Support**: English-only interface initially (internationalization not in scope)
9. **Mobile App**: Web-based responsive interface only (no native iOS/Android apps)
10. **Browser Support**: Modern browsers only (last 2 versions of Chrome, Firefox, Safari, Edge)
11. **Time-Series Database**: Assumes TimescaleDB (PostgreSQL extension) or ClickHouse deployment for analytics data
12. **Snowflake ID Generator**: Assumes Snowflake ID generator infrastructure (worker ID coordination, time synchronization)
13. **Redis/Valkey Cluster**: Assumes Redis or Valkey cluster deployment for L2 distributed caching

## Out of Scope

The following items are explicitly excluded from this specification:

1. **Billing/Monetization**: No payment processing, premium tiers, or subscription management
2. **Link Expiration Automation**: No scheduled jobs to auto-expire links (users set expiration but manual enforcement)
3. **QR Code Generation**: No QR code creation for short URLs (may be future enhancement)
4. **Branded Domains**: Users cannot use custom domains (e.g., "links.mybrand.com") - single shared domain only
5. **Team/Organization Accounts**: Individual accounts only, no shared workspaces or team analytics
6. **Advanced Analytics**: No A/B testing, conversion tracking, or integration with Google Analytics/analytics platforms
7. **Link Preview**: No Open Graph preview/thumbnail generation for short URLs
8. **Browser Extensions**: No Chrome/Firefox extensions for quick shortening
9. **Email Campaigns**: No bulk email sending or newsletter integration
10. **Social Media Integration**: No auto-posting shortened links to Twitter/Facebook/LinkedIn
11. **Password Policies**: No enforced complexity rules (beyond minimum length)
12. **Two-Factor Authentication (2FA)**: Not included in initial release
13. **Audit Logs**: No detailed activity logs beyond basic analytics
14. **Export Features**: No PDF export of analytics data (CSV export included for basic data portability)
15. **Webhook Notifications**: No real-time notifications for click events

