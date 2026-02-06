# ELK Tutorial - Search & Analytics Mastery

Learn log aggregation with Elasticsearch, Kibana, and KQL/Lucene.

## Quick Start

```bash
task kind:conduct -- elk-tutorial
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         KIBANA                              │
│                    (Visualization)                          │
│                          │                                  │
│                    KQL/Lucene queries                       │
│                          ▼                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   ELASTICSEARCH                      │   │
│  │              (Log Storage & Search)                  │   │
│  │                                                      │   │
│  │  • Full-text indexed (inverted index)                │   │
│  │  • Documents in shards                               │   │
│  │  • KQL and Lucene query languages                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                          ▲                                  │
│                     Push logs                               │
│                          │                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    FLUENT BIT                        │   │
│  │               (Log Collection)                       │   │
│  │                                                      │   │
│  │  • DaemonSet on each node                            │   │
│  │  • Tails container logs                              │   │
│  │  • Adds Kubernetes metadata                          │   │
│  └─────────────────────────────────────────────────────┘   │
│                          ▲                                  │
│                    Container logs                           │
│                          │                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  LOG GENERATOR                       │   │
│  │              (Demo Application)                      │   │
│  │                                                      │   │
│  │  • Structured JSON logs                              │   │
│  │  • Configurable rate & cardinality                   │   │
│  │  • Multiple log levels (debug/info/warn/error)       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## What You'll Learn

### Module 1: KQL (Kibana Query Language)

KQL is Kibana's default query syntax - simple and intuitive.

```kql
# Field search
level:error

# Wildcard
service:service-*

# Numeric range
status >= 500

# Multiple conditions (AND is implicit)
level:error service:service-1

# OR condition
level:error OR level:warn

# NOT condition
level:error AND NOT service:service-1

# Nested field
kubernetes.labels.app:logging-app
```

### Module 2: Lucene Syntax

Lucene is more powerful for complex searches.

```lucene
# Phrase search (exact match)
message:"connection failed"

# Wildcard
trace_id:abc*

# Range queries
duration_ms:[1000 TO 5000]

# Fuzzy search (typo tolerance)
message:databse~

# Boolean with parentheses
(level:error OR level:warn) AND service:service-1

# Exists query
_exists_:trace_id
```

### Module 3: Aggregations

Elasticsearch aggregations power Kibana visualizations.

```json
// Terms aggregation (top values)
{
  "aggs": {
    "by_level": {
      "terms": { "field": "level" }
    }
  }
}

// Date histogram (over time)
{
  "aggs": {
    "over_time": {
      "date_histogram": {
        "field": "@timestamp",
        "fixed_interval": "1m"
      }
    }
  }
}

// Percentiles
{
  "aggs": {
    "duration_percentiles": {
      "percentiles": {
        "field": "duration_ms",
        "percents": [50, 95, 99]
      }
    }
  }
}
```

### Module 4: Index Templates & Mappings

Control how Elasticsearch indexes your data.

```json
// Index template for logs
{
  "index_patterns": ["logs-*"],
  "template": {
    "mappings": {
      "properties": {
        "level": { "type": "keyword" },
        "message": { "type": "text" },
        "duration_ms": { "type": "integer" },
        "@timestamp": { "type": "date" }
      }
    }
  }
}
```

Field types:
- **keyword**: Exact match, aggregatable (level, service)
- **text**: Full-text searchable (message)
- **integer/long**: Numeric (status, duration_ms)
- **date**: Timestamps (@timestamp)

### Module 5: Index Lifecycle Management (ILM)

Manage data retention automatically.

```json
// ILM Policy
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_age": "1d",
            "max_primary_shard_size": "10gb"
          }
        }
      },
      "warm": {
        "min_age": "7d",
        "actions": {
          "shrink": { "number_of_shards": 1 }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
```

## Kibana Navigation

| Section | Purpose |
|---------|---------|
| **Discover** | Search and explore logs |
| **Dashboard** | Create and view dashboards |
| **Visualize** | Build individual visualizations |
| **Stack Management** | Configure data views, ILM, security |

## Log Generator Output

Same format as loki-tutorial for fair comparison:

```json
{
  "timestamp": "2024-01-15T10:23:45.123456789Z",
  "level": "error",
  "service": "service-3",
  "endpoint": "/api/v1/orders",
  "method": "POST",
  "status": 500,
  "duration_ms": 1234,
  "trace_id": "abc123def456...",
  "message": "Database connection failed"
}
```

## Elasticsearch vs Loki

| Aspect | Elasticsearch | Loki |
|--------|---------------|------|
| **Indexing** | Full content | Labels only |
| **Query speed (labels)** | Fast | Fast |
| **Query speed (content)** | Fast (inverted index) | Slow (grep) |
| **Storage** | Large (indexes everything) | 10-100x smaller |
| **RAM usage** | ~2GB+ (JVM) | ~300MB |
| **Best for** | Ad-hoc search, analytics | K8s debugging, known patterns |

## API Examples

```bash
# Cluster health
curl "http://$ES_IP:9200/_cluster/health?pretty"

# List indices
curl "http://$ES_IP:9200/_cat/indices?v"

# Search with query
curl -X POST "http://$ES_IP:9200/logs-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": { "match": { "level": "error" } },
    "size": 10
  }'

# Aggregation
curl -X POST "http://$ES_IP:9200/logs-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 0,
    "aggs": {
      "by_service": {
        "terms": { "field": "service" }
      }
    }
  }'
```

## Next Steps

- **loki-tutorial**: Learn Grafana Loki + LogQL for comparison
- **logging-comparison**: Run both stacks side-by-side

## References

- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [KQL Reference](https://www.elastic.co/guide/en/kibana/current/kuery-query.html)
- [Lucene Query Syntax](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html#query-string-syntax)
- [ILM Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-lifecycle-management.html)
- [ADR-010: Logging Stack Selection](../../docs/adrs/ADR-010-logging-stack-selection.md)
- [ADR-011: Observability Architecture](../../docs/adrs/ADR-011-observability-architecture.md)
