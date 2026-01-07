# SeaweedFS Tutorial: Needle in a Haystack

Learn why Facebook invented Haystack and how SeaweedFS achieves O(1) file lookups.

## The Problem

Traditional filesystems struggle with millions of small files:

```
Reading 1 file from 10 files:      ~1ms
Reading 1 file from 10,000 files:  ~5ms
Reading 1 file from 1,000,000 files: ~50ms+ (directory scanning)
```

This is O(n) - lookup time grows with file count.

## The Haystack Solution

Facebook's 2010 paper introduced a simple insight:

> "Pack millions of small files into large volumes. Keep the index in memory."

```
Traditional:                    Haystack/SeaweedFS:
─────────────                   ──────────────────
/photos/                        volume_001.dat (32GB)
  photo_001.jpg                   [photo_001 | photo_002 | photo_003 | ...]
  photo_002.jpg
  photo_003.jpg                 volume_001.idx (in memory)
  ... (millions)                  001 -> offset 0
                                  002 -> offset 4096
                                  003 -> offset 8192
```

Lookup: file_id → volume_id → offset → **single disk read** = O(1)

## Tutorial Scenario

The starship USS Kubernetes has 10 years of sensor data (100,000 readings).
During an emergency, you need to find a specific reading from stardate 47634.44.

With traditional storage: minutes of searching.
With SeaweedFS: instant retrieval.

## Quick Start

```bash
# Run the experiment
task kind:conduct -- seaweedfs-tutorial
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SeaweedFS Cluster                        │
│                                                             │
│  ┌──────────────┐     ┌──────────────┐     ┌─────────────┐ │
│  │    Master    │     │   Volume     │     │    Filer    │ │
│  │   (9333)     │────▶│   Server     │◀───▶│   (8888)    │ │
│  │              │     │   (8080)     │     │             │ │
│  │ - Volume map │     │              │     │ - S3 API    │ │
│  │ - Topology   │     │ - .dat files │     │ - REST API  │ │
│  └──────────────┘     │ - .idx files │     │ - POSIX     │ │
│                       └──────────────┘     └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Checkpoints

1. **Deploy SeaweedFS** - Master, Volume, Filer running
2. **Explore Architecture** - Access master UI at :9333
3. **Load Blackbox** - Upload 100,000 sensor readings
4. **Find the Needle** - Retrieve stardate 47634.44 instantly
5. **LOSF Problem** - See why filesystems fail at scale
6. **Volume Internals** - Examine .dat and .idx files
7. **S3 Gateway** - Use AWS CLI with SeaweedFS
8. **Create Buckets** - Prepare for Loki/Thanos integration

## Key Commands

```bash
# Access SeaweedFS
export FILER_IP=$(kubectl get svc -n seaweedfs seaweedfs-filer-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export S3_IP=$(kubectl get svc -n seaweedfs seaweedfs-s3-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export MASTER_IP=$(kubectl get svc -n seaweedfs seaweedfs-master-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Find the needle (O(1) lookup!)
time curl http://$FILER_IP:8888/blackbox/sensors/47634.44.json

# List all sensor readings
curl http://$FILER_IP:8888/blackbox/sensors/?limit=10

# Use S3 API
aws --endpoint-url http://$S3_IP:8333 s3 ls

# View volume status
curl http://$MASTER_IP:9333/dir/status

# Run benchmark
kubectl exec -n seaweedfs deploy/blackbox-loader -- /scripts/benchmark.sh
```

## Learning Outcomes

After this tutorial, you'll understand:

- Why billions of small files break traditional filesystems
- How Haystack/SeaweedFS achieves O(1) lookups
- The master/volume/filer architecture
- S3 API compatibility for tool integration
- When to use SeaweedFS vs cloud S3

## Next Steps

- [Loki Tutorial](../loki-tutorial/) - Use SeaweedFS for log storage
- [Thanos Tutorial](../thanos-tutorial/) - Use SeaweedFS for metrics
