use std::net::IpAddr;

use anyhow::Result;
use maxminddb::{Reader, geoip2};
use tracing::warn;

use crate::geo::GeoInfo;

pub struct GeoLookup {
    reader: maxminddb::Reader<Vec<u8>>,
}

impl GeoLookup {
    pub fn new(db_path: &str) -> Result<Self> {
        let reader = Reader::open_readfile(db_path)?;
        Ok(Self { reader })
    }

    pub fn lookup(&self, ip: IpAddr) -> Option<GeoInfo> {
        let result = match self.reader.lookup(ip) {
            Ok(r) => r,
            Err(e) => {
                if !is_private_ip(ip) {
                    warn!(ip = %ip, error = %e, "GeoIP lookup failed");
                }
                return None;
            }
        };

        let city_data: geoip2::City = match result.decode() {
            Ok(Some(c)) => c,
            Ok(None) => return None,
            Err(e) => {
                warn!(ip = %ip, error = %e, "GeoIP decode failed");
                return None;
            }
        };

        Some(GeoInfo {
            country_code: city_data.country.iso_code.map(|s| s.to_string()),
            country_name: city_data.country.names.english.map(|s| s.to_string()),
            city: city_data.city.names.english.map(|s| s.to_string()),
            latitude: city_data.location.latitude,
            longitude: city_data.location.longitude,
        })
    }
}

fn is_private_ip(ip: IpAddr) -> bool {
    match ip {
        IpAddr::V4(v4) => {
            v4.is_private() || v4.is_loopback() || v4.is_link_local() || v4.is_unspecified()
        }
        IpAddr::V6(v6) => v6.is_loopback() || v6.is_unspecified(),
    }
}
