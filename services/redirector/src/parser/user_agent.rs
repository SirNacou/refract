use std::fs::File;

use anyhow::Result;
use ua_parser::{Extractor, Regexes};

#[derive(Debug, Clone, Default)]
pub struct ParsedUserAgent {
    pub device_type: String,
    pub browser: Option<String>,
    pub operating_system: Option<String>,
}

pub struct UAParser {
    extractor: Extractor<'static>,
}

impl UAParser {
    pub fn new(yaml_path: &str) -> Result<Self> {
        let file = File::open(yaml_path)?;
        let regexes: Regexes<'static> = serde_yaml::from_reader(file)?;
        let extractor = Extractor::try_from(regexes)?;
        Ok(Self { extractor })
    }

    pub fn parse(&self, user_agent: &str) -> ParsedUserAgent {
        let (ua, os, device) = self.extractor.extract(user_agent);

        let device_type = classify_device_type(
            device.as_ref().map(|d| d.device.as_ref()),
            ua.as_ref().map(|u| u.family.as_ref()),
        );

        let browser = ua.as_ref().and_then(format_browser);

        let operating_system = os.as_ref().and_then(format_os);

        ParsedUserAgent {
            device_type,
            browser,
            operating_system,
        }
    }
}

fn classify_device_type(device: Option<&str>, ua_family: Option<&str>) -> String {
    let device_lower = device.unwrap_or("").to_lowercase();
    let ua_lower = ua_family.unwrap_or("").to_lowercase();

    // Bot detection (check UA family - more reliable for crawlers)
    if ua_lower.contains("bot")
        || ua_lower.contains("spider")
        || ua_lower.contains("crawler")
        || ua_lower.contains("googlebot")
        || ua_lower.contains("bingbot")
        || device_lower.contains("spider")
        || device_lower.contains("bot")
    {
        return "bot".to_string();
    }

    // Tablet detection
    if device_lower.contains("ipad")
        || device_lower.contains("tablet")
        || device_lower.contains("kindle")
        || device_lower.contains("playbook")
    {
        return "tablet".to_string();
    }

    // Mobile detection
    if device_lower.contains("iphone")
        || device_lower.contains("android")
        || device_lower.contains("mobile")
        || device_lower.contains("phone")
        || device_lower.contains("ipod")
    {
        return "mobile".to_string();
    }

    // Default to desktop
    "desktop".to_string()
}
fn format_browser(ua: &ua_parser::user_agent::ValueRef<'_>) -> Option<String> {
    // Skip if family is empty or "Other"
    if ua.family.is_empty() || ua.family == "Other" {
        return None;
    }

    let mut result = ua.family.to_string();

    if let Some(major) = ua.major {
        result.push(' ');
        result.push_str(major);
        if let Some(minor) = ua.minor {
            result.push('.');
            result.push_str(minor);
        }
    }

    Some(result)
}
fn format_os(os: &ua_parser::os::ValueRef<'_>) -> Option<String> {
    // Skip if os is empty or "Other"
    if os.os.is_empty() || os.os == "Other" {
        return None;
    }

    let mut result = os.os.to_string();

    if let Some(ref major) = os.major {
        result.push(' ');
        result.push_str(major);
        if let Some(ref minor) = os.minor {
            result.push('.');
            result.push_str(minor);
        }
    }

    Some(result)
}
