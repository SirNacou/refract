use std::net::{IpAddr, Ipv4Addr, Ipv6Addr};

use jiff::Timestamp;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

pub mod publisher;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClickEvent {
    pub event_id: Uuid,
    pub url_id: i64,
    pub short_code: String,
    pub timestamp: Timestamp,
    pub user_agent: String,
    pub ip_address: IpAddr,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub referrer: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub country_code: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub country_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub city: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub latitude: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub longitude: Option<f64>,
    pub device_type: String, // desktop, mobile, tablet, bot
    #[serde(skip_serializing_if = "Option::is_none")]
    pub browser: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub operating_system: Option<String>,
    pub cache_tier: String, // l1, l2, db
    pub latency_ms: f64,
    pub request_id: String,
}

impl ClickEvent {
    pub fn new(
        url_id: i64,
        short_code: String,
        user_agent: String,
        ip_address: IpAddr,
        referrer: Option<String>,
        country_code: Option<String>,
        country_name: Option<String>,
        city: Option<String>,
        latitude: Option<f64>,
        longitude: Option<f64>,
        device_type: String,
        browser: Option<String>,
        operating_system: Option<String>,
        cache_tier: String,
        latency_ms: f64,
        request_id: String,
    ) -> Self {
        let event_id = Uuid::now_v7();
        let (s, n) = event_id.get_timestamp().unwrap().to_unix();
        let timestamp = Timestamp::new(s.try_into().unwrap(), n.try_into().unwrap()).unwrap();
        Self {
            event_id,
            url_id,
            short_code,
            timestamp,
            user_agent,
            ip_address,
            referrer,
            country_code,
            country_name,
            city,
            latitude,
            longitude,
            device_type,
            browser,
            operating_system,
            cache_tier,
            latency_ms,
            request_id,
        }
    }

    /// Anonymize IP: zero last octet for IPv4, last 80 bits for IPv6
    pub fn anonymize_ip(ip: IpAddr) -> IpAddr {
        match ip {
            IpAddr::V4(v4) => {
                let octets = v4.octets();
                IpAddr::V4(Ipv4Addr::new(octets[0], octets[1], octets[2], 0))
            }
            IpAddr::V6(v6) => {
                let mut segments = v6.segments();
                // Zero last 5 segments (80 bits) per spec
                segments[3] = 0;
                segments[4] = 0;
                segments[5] = 0;
                segments[6] = 0;
                segments[7] = 0;
                IpAddr::V6(Ipv6Addr::from(segments))
            }
        }
    }
}
