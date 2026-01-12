use std::sync::Arc;

use anyhow::Result;
use multi_tier_cache::{CacheManager, CacheSystem, CacheSystemBuilder, L1Cache, RedisCache};

use crate::config::RedisConfig;

pub struct RedirectCache {
    cache: CacheSystem,
}

impl RedirectCache {
    pub async fn new(cfg: &RedisConfig) -> Result<Self> {
        let redis = RedisCache::with_url(&cfg.to_redis_url()).await?;
        let redis = Arc::new(redis);
        let cache = CacheSystemBuilder::new()
            .with_l1(Arc::new(L1Cache::new(
                multi_tier_cache::MokaCacheConfig::default(),
            )?))
            .with_l2(redis)
            .build()
            .await?;
        Ok(Self { cache })
    }

    pub fn get_cache_manager(&self) -> &Arc<CacheManager> {
        self.cache.cache_manager()
    }
}
