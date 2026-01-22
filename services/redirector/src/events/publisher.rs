use std::{sync::Arc, time::Duration};

use anyhow::Result;
use multi_tier_cache::{CacheManager, RedisStreams};
use tokio::sync::Mutex;
use tracing::{debug, info, warn};

use crate::{
    config::{EventsConfig, RedisConfig},
    events::ClickEvent,
};

/// Publishes click events to Redis Stream asynchronously
///
/// # Design
/// - Events buffered in memory (non-blocking publish)
/// - Flushed when batch_size reached OR flush_interval expires
/// - Buffer overflow protection: drops oldest when max_buffer_size exceeded
/// - Graceful degradation: Redis failures logged, not propagated
pub struct ClickEventPublisher {
    redis_stream: Arc<RedisStreams>,
    stream_key: String,
    max_stream_len: usize,
    buffer: Arc<Mutex<Vec<ClickEvent>>>,
    batch_size: usize,
    flush_interval_ms: u64,
    max_buffer_size: usize,
}
impl ClickEventPublisher {
    /// Create a new publisher connected to Redis Streams
    ///
    /// # Arguments
    /// * `redis_cfg` - Redis connection configuration
    /// * `events_cfg` - Event publishing configuration
    ///
    /// # Errors
    /// Returns error if Redis connection fails
    pub async fn new(redis_config: &RedisConfig, events_cfg: &EventsConfig) -> Result<Self> {
        info!(
            stream_key = %events_cfg.stream_key,
            batch_size = events_cfg.batch_size,
            flush_interval_ms = events_cfg.flush_interval_ms,
            max_buffer_size = events_cfg.max_buffer_size,
            "ClickEventPublisher initialized"
        );

        let stream = Arc::new(RedisStreams::new(&redis_config.to_redis_url()).await?);

        Ok(Self {
            redis_stream: stream,
            stream_key: events_cfg.stream_key.clone(),
            max_stream_len: events_cfg.max_stream_len,
            buffer: Arc::new(Mutex::new(Vec::with_capacity(events_cfg.batch_size))),
            batch_size: events_cfg.batch_size,
            flush_interval_ms: events_cfg.flush_interval_ms,
            max_buffer_size: events_cfg.max_buffer_size,
        })
    }
    /// Publish a click event (non-blocking)
    ///
    /// Adds event to buffer. If buffer reaches batch_size, triggers async flush.
    /// If buffer exceeds max_buffer_size, oldest events are dropped.
    pub fn publish(self: &Arc<Self>, event: ClickEvent) {
        let publisher = self.clone();
        tokio::spawn(async move {
            let should_flush = {
                let mut buf = publisher.buffer.lock().await;
                // Buffer overflow protection: drop oldest events
                if buf.len() >= publisher.max_buffer_size {
                    let drop_count = buf.len() - publisher.max_buffer_size + 1;
                    warn!(
                        drop_count = drop_count,
                        max_buffer_size = publisher.max_buffer_size,
                        "Buffer overflow, dropping oldest events"
                    );
                    buf.drain(0..drop_count);
                }
                buf.push(event);
                buf.len() >= publisher.batch_size
            };
            if should_flush {
                if let Err(e) = publisher.flush().await {
                    warn!(error = %e, "Failed to flush click events to Redis");
                }
            }
        });
    }
    /// Start background task that flushes buffer periodically
    ///
    /// Call this once after creating the publisher.
    pub fn start_flush_task(self: &Arc<Self>) {
        let publisher = self.clone();
        tokio::spawn(async move {
            let mut interval =
                tokio::time::interval(Duration::from_millis(publisher.flush_interval_ms));
            loop {
                interval.tick().await;
                if let Err(e) = publisher.flush().await {
                    warn!(error = %e, "Periodic flush failed");
                }
            }
        });
    }
    /// Flush all buffered events to Redis Stream
    ///
    /// Called automatically by publish() and the background task.
    /// Can also be called manually for graceful shutdown.
    pub async fn flush(&self) -> Result<()> {
        // Take all events from buffer (swap with empty vec)
        let events: Vec<ClickEvent> = {
            let mut buf = self.buffer.lock().await;
            std::mem::take(&mut *buf)
        };
        if events.is_empty() {
            return Ok(());
        }
        let count = events.len();
        debug!(count = count, "Flushing click events to Redis Stream");
        for event in events {
            // Serialize entire event to JSON
            let json = serde_json::to_string(&event)?;
            self.redis_stream
                .stream_add(
                    &self.stream_key,
                    vec![("data".to_string(), json)],
                    Some(self.max_stream_len),
                )
                .await?;
        }
        debug!(count = count, "Successfully flushed click events");
        Ok(())
    }
}
