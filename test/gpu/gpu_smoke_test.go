// Copyright 2023 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package gpu_smoke_test tests basic GPU functionality.
package gpu_smoke_test

import (
	"context"
	"testing"
	"time"

	"gvisor.dev/gvisor/pkg/test/dockerutil"
	"gvisor.dev/gvisor/pkg/test/testutil"
)

func TestGPUHello(t *testing.T) {
	ctx := context.Background()
	c := dockerutil.MakeContainer(ctx, t)
	defer c.CleanUp(ctx)

	opts := dockerutil.GPURunOpts()
	opts.Image = "basic/cuda-vector-add"
	out, err := c.Run(ctx, opts)
	if err != nil {
		t.Fatalf("could not run cuda-vector-add: %v", err)
	}
	t.Logf("cuda-vector-add output: %s", string(out))
}

func TestGPUCheckpointRestore(t *testing.T) {
	if !testutil.IsCheckpointSupported() {
		t.Skip("Checkpoint is not supported.")
	}
	dockerutil.EnsureDockerExperimentalEnabled()

	ctx := context.Background()
	c := dockerutil.MakeContainer(ctx, t)
	defer c.CleanUp(ctx)

	opts := dockerutil.GPURunOpts()
	opts.Image = "basic/cuda-vector-add"
	if err := c.Spawn(ctx, opts, "sleep", "infinity"); err != nil {
		t.Fatalf("could not run cuda-vector-add: %v", err)
	}

	// Run the vector add program.
	if _, err := c.Exec(ctx, dockerutil.ExecOpts{}, "/bin/sh", "-c", "./vectorAdd"); err != nil {
		t.Fatalf("docker exec failed: %v", err)
	}

	// Create a snapshot.
	if err := c.Checkpoint(ctx, "test"); err != nil {
		t.Fatalf("docker checkpoint failed: %v", err)
	}
	if err := c.WaitTimeout(ctx, time.Minute); err != nil {
		t.Fatalf("wait failed: %v", err)
	}

	// Restore the snapshot.
	// TODO(b/143498576): Remove Poll after github.com/moby/moby/issues/38963 is fixed.
	if err := testutil.Poll(func() error { return c.Restore(ctx, "test") }, time.Minute); err != nil {
		t.Fatalf("docker restore failed: %v", err)
	}

	// Run the vector add program again to ensure GPUs are functional.
	if _, err := c.Exec(ctx, dockerutil.ExecOpts{}, "/bin/sh", "-c", "./vectorAdd"); err != nil {
		t.Fatalf("docker exec failed: %v", err)
	}
}

func TestCUDATests(t *testing.T) {
	ctx := context.Background()
	c := dockerutil.MakeContainer(ctx, t)
	defer c.CleanUp(ctx)

	opts := dockerutil.GPURunOpts()
	opts.Image = "gpu/cuda-tests"
	out, err := c.Run(ctx, opts)
	if err != nil {
		t.Fatalf("could not run cuda-tests: %v", err)
	}
	t.Logf("cuda-tests output: %s", string(out))
}
