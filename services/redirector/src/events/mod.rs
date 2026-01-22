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
}

impl ClickEvent {
    pub fn new(
        url_id: i64,
        short_code: &str,
        user_agent: &str,
        ip_address: IpAddr,
        referrer: Option<String>,
    ) -> Self {
        let event_id = Uuid::now_v7();
        let (s, n) = event_id.get_timestamp().unwrap().to_unix();
        let timestamp = Timestamp::new(s.try_into().unwrap(), n.try_into().unwrap()).unwrap();
        Self {
            event_id,
            url_id,
            short_code: short_code.to_string(),
            timestamp,
            user_agent: user_agent.to_string(),
            ip_address,
            referrer,
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
