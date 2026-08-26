// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package reference

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed reference.wasm
var wasm []byte

//go:embed reference.wasm.sha256
var wasmChecksum []byte

const (
	referenceABIVersion       uint64 = 1
	referenceMemoryLimitPages uint32 = 1_024
)

type (
	Object = uint32

	Engine struct {
		module api.Module
	}
)

var (
	runtimeOnce     sync.Once
	runtimeInstance wazero.Runtime
	compiledModule  wazero.CompiledModule
	runtimeErr      error
	moduleSequence  atomic.Uint64
)

func New() (*Engine, error) {
	backend, err := instantiate()
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate Automerge reference backend: %w", err)
	}

	if err := backend.run("am_create"); err != nil {
		_ = backend.Close()
		return nil, fmt.Errorf("cannot create Automerge document: %w", err)
	}

	return backend, nil
}

func Load(document []byte) (*Engine, error) {
	backend, err := instantiate()
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate Automerge reference backend: %w", err)
	}

	if err := backend.runBytes("am_load", document); err != nil {
		_ = backend.Close()
		return nil, fmt.Errorf("cannot load Automerge document: %w", err)
	}

	return backend, nil
}

// LoadConvertingStrings loads a document, converting every string scalar in a
// map or list into a text object, mirroring StringMigration::ConvertToText.
func LoadConvertingStrings(document []byte) (*Engine, error) {
	backend, err := instantiate()
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate Automerge reference backend: %w", err)
	}

	if err := backend.runBytes("am_load_convert_strings", document); err != nil {
		_ = backend.Close()

		return nil, fmt.Errorf("cannot load Automerge document with string migration: %w", err)
	}

	return backend, nil
}

func instantiate() (*Engine, error) {
	runtimeOnce.Do(
		func() {
			fields := strings.Fields(string(wasmChecksum))
			if len(fields) == 0 {
				runtimeErr = fmt.Errorf("automerge reference checksum is empty")
				return
			}

			checksum := sha256.Sum256(wasm)

			actualChecksum := hex.EncodeToString(checksum[:])
			if !strings.EqualFold(fields[0], actualChecksum) {
				runtimeErr = fmt.Errorf(
					"automerge reference checksum mismatch: expected %s, got %s",
					fields[0],
					actualChecksum,
				)

				return
			}

			runtimeInstance = wazero.NewRuntimeWithConfig(
				context.Background(),
				wazero.NewRuntimeConfig().WithMemoryLimitPages(referenceMemoryLimitPages),
			)
			if _, err := wasi_snapshot_preview1.Instantiate(context.Background(), runtimeInstance); err != nil {
				runtimeErr = fmt.Errorf("cannot instantiate WASI: %w", err)
				return
			}

			compiledModule, runtimeErr = runtimeInstance.CompileModule(context.Background(), wasm)
			if runtimeErr != nil {
				runtimeErr = fmt.Errorf("cannot compile reference module: %w", runtimeErr)
			}
		},
	)

	if runtimeErr != nil {
		return nil, runtimeErr
	}

	name := fmt.Sprintf("automerge-reference-%d", moduleSequence.Add(1))

	module, err := runtimeInstance.InstantiateModule(
		context.Background(),
		compiledModule,
		wazero.NewModuleConfig().WithName(name).WithRandSource(rand.Reader),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate reference module: %w", err)
	}

	backend := &Engine{module: module}

	version, err := backend.call("am_abi_version")
	if err != nil {
		_ = module.Close(context.Background())
		return nil, fmt.Errorf("cannot read reference ABI version: %w", err)
	}

	if version[0] != referenceABIVersion {
		_ = module.Close(context.Background())

		return nil, fmt.Errorf(
			"unsupported reference ABI version %d, expected %d",
			version[0],
			referenceABIVersion,
		)
	}

	return backend, nil
}

func (b *Engine) Close() error {
	if err := b.module.Close(context.Background()); err != nil {
		return fmt.Errorf("cannot close reference module: %w", err)
	}

	return nil
}

// Save serializes the document. retainOrphans keeps changes whose dependencies
// are missing; compress DEFLATEs the output. The reference exposes three save
// entry points, so the flag combination maps onto the closest one: the default
// am_save (retain, compress), am_save_nocompress (retain, no compress), and
// am_save_no_orphans (discard orphans, which also does not compress).
func (b *Engine) Save(
	retainOrphans bool,
	compress bool,
) ([]byte, error) {
	function := "am_save"

	switch {
	case !retainOrphans:
		function = "am_save_no_orphans"
	case !compress:
		function = "am_save_nocompress"
	}

	if err := b.run(function); err != nil {
		return nil, fmt.Errorf("cannot save reference document: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot read saved reference document: %w", err)
	}

	return output, nil
}

func (b *Engine) SaveIncremental() ([]byte, error) {
	if err := b.run("am_save_incremental"); err != nil {
		return nil, fmt.Errorf("cannot save incremental reference changes: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy incremental reference changes: %w", err)
	}

	return output, nil
}

func (b *Engine) LoadIncremental(
	data []byte,
) (uint64, error) {
	pointer, length, err := b.write(data)
	if err != nil {
		return 0, fmt.Errorf("cannot write incremental reference changes: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call(
		"am_load_incremental",
		uint64(pointer),
		uint64(length),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot load incremental reference changes: %w", err)
	}

	applied := int64(result[0])
	if applied < 0 {
		return 0, b.operationError("cannot load incremental reference changes")
	}

	return uint64(applied), nil
}

// Isolate pins the document to the given heads, mirroring AutoCommit::isolate.
func (b *Engine) Isolate(heads [][32]byte) error {
	if err := b.runBytes("am_isolate", flattenHashes(heads)); err != nil {
		return fmt.Errorf("cannot isolate reference document: %w", err)
	}

	return nil
}

// Integrate ends isolation, mirroring AutoCommit::integrate.
func (b *Engine) Integrate() error {
	if err := b.run("am_integrate"); err != nil {
		return fmt.Errorf("cannot integrate reference document: %w", err)
	}

	return nil
}

func (b *Engine) SetActor(actor []byte) error {
	if err := b.runBytes("am_set_actor", actor); err != nil {
		return fmt.Errorf("cannot set reference actor: %w", err)
	}

	return nil
}

func (b *Engine) PutString(object Object, key, value string) error {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return fmt.Errorf("cannot write map key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	valuePointer, valueLength, err := b.write([]byte(value))
	if err != nil {
		return fmt.Errorf("cannot write map value: %w", err)
	}
	defer b.free(valuePointer, valueLength)

	if err := b.run(
		"am_put_string",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
		uint64(valuePointer),
		uint64(valueLength),
	); err != nil {
		return fmt.Errorf("cannot put reference map value: %w", err)
	}

	return nil
}

func (b *Engine) GetString(object Object, key string) (string, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return "", fmt.Errorf("cannot write map key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	if err := b.run(
		"am_get_string",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
	); err != nil {
		return "", fmt.Errorf("cannot get reference map value: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return "", fmt.Errorf("cannot copy reference map value: %w", err)
	}

	return string(output), nil
}

func (b *Engine) PutScalar(
	object Object,
	key string,
	value []byte,
) error {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return fmt.Errorf("cannot write scalar key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	valuePointer, valueLength, err := b.write(value)
	if err != nil {
		return fmt.Errorf("cannot write scalar value: %w", err)
	}
	defer b.free(valuePointer, valueLength)

	if err := b.run(
		"am_put_scalar",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
		uint64(valuePointer),
		uint64(valueLength),
	); err != nil {
		return fmt.Errorf("cannot put reference scalar: %w", err)
	}

	return nil
}

func (b *Engine) GetScalar(
	object Object,
	key string,
) ([]byte, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("cannot write scalar key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	if err := b.run(
		"am_get_scalar",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
	); err != nil {
		return nil, fmt.Errorf("cannot get reference scalar: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference scalar: %w", err)
	}

	return output, nil
}

func (b *Engine) GetScalarAtHeads(
	object Object,
	key string,
	heads [][32]byte,
) ([]byte, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("cannot write historical scalar key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	headPointer, headLength, err := b.write(flattenHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot write historical scalar heads: %w", err)
	}
	defer b.free(headPointer, headLength)

	if err := b.run(
		"am_get_scalar_at_heads",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
		uint64(headPointer),
		uint64(headLength),
	); err != nil {
		return nil, fmt.Errorf("cannot get historical reference scalar: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy historical reference scalar: %w", err)
	}

	return output, nil
}

func (b *Engine) GetAllScalars(
	object Object,
	key string,
) ([]byte, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("cannot write scalar key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	if err := b.run(
		"am_get_all_scalars",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
	); err != nil {
		return nil, fmt.Errorf("cannot get reference scalar conflicts: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference scalar conflicts: %w", err)
	}

	return output, nil
}

func (b *Engine) GetAllScalarsAt(
	object Object,
	index uint64,
) ([]byte, error) {
	if err := b.run(
		"am_get_all_scalars_at",
		uint64(object),
		index,
	); err != nil {
		return nil, fmt.Errorf("cannot get reference sequence scalar conflicts: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference sequence scalar conflicts: %w", err)
	}

	return output, nil
}

func (b *Engine) PutObject(
	object Object,
	key string,
	objectType string,
) (Object, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return 0, fmt.Errorf("cannot write object key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	typePointer, typeLength, err := b.write([]byte(objectType))
	if err != nil {
		return 0, fmt.Errorf("cannot write object type: %w", err)
	}
	defer b.free(typePointer, typeLength)

	result, err := b.call(
		"am_put_object",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
		uint64(typePointer),
		uint64(typeLength),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot create reference object: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot create reference object")
	}

	return Object(handle), nil
}

func (b *Engine) GetObject(
	object Object,
	key string,
) (Object, string, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return 0, "", fmt.Errorf("cannot write object key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	result, err := b.call(
		"am_get_object",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
	)
	if err != nil {
		return 0, "", fmt.Errorf("cannot get reference object: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, "", b.operationError("cannot get reference object")
	}

	rawType, err := b.output()
	if err != nil {
		return 0, "", fmt.Errorf("cannot copy reference object type: %w", err)
	}

	return Object(handle), string(rawType), nil
}

func (b *Engine) InsertScalar(
	object Object,
	index uint64,
	value []byte,
) error {
	pointer, length, err := b.write(value)
	if err != nil {
		return fmt.Errorf("cannot write sequence scalar: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_insert_scalar",
		uint64(object),
		index,
		uint64(pointer),
		uint64(length),
	); err != nil {
		return fmt.Errorf("cannot insert reference scalar: %w", err)
	}

	return nil
}

func (b *Engine) PutScalarAt(
	object Object,
	index uint64,
	value []byte,
) error {
	pointer, length, err := b.write(value)
	if err != nil {
		return fmt.Errorf("cannot write sequence scalar: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_put_scalar_at",
		uint64(object),
		index,
		uint64(pointer),
		uint64(length),
	); err != nil {
		return fmt.Errorf("cannot replace reference scalar: %w", err)
	}

	return nil
}

func (b *Engine) InsertObject(
	object Object,
	index uint64,
	objectType string,
) (Object, error) {
	pointer, length, err := b.write([]byte(objectType))
	if err != nil {
		return 0, fmt.Errorf("cannot write sequence object type: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call(
		"am_insert_object",
		uint64(object),
		index,
		uint64(pointer),
		uint64(length),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot insert reference object: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot insert reference object")
	}

	return Object(handle), nil
}

func (b *Engine) PutObjectAt(
	object Object,
	index uint64,
	objectType string,
) (Object, error) {
	pointer, length, err := b.write([]byte(objectType))
	if err != nil {
		return 0, fmt.Errorf("cannot write replacement object type: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call(
		"am_put_object_at",
		uint64(object),
		index,
		uint64(pointer),
		uint64(length),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot replace reference object: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot replace reference object")
	}

	return Object(handle), nil
}

func (b *Engine) GetScalarAt(
	object Object,
	index uint64,
) ([]byte, error) {
	if err := b.run(
		"am_get_scalar_at",
		uint64(object),
		index,
	); err != nil {
		return nil, fmt.Errorf("cannot get reference sequence scalar: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference sequence scalar: %w", err)
	}

	return output, nil
}

func (b *Engine) GetObjectAt(
	object Object,
	index uint64,
) (Object, string, error) {
	result, err := b.call(
		"am_get_object_at",
		uint64(object),
		index,
	)
	if err != nil {
		return 0, "", fmt.Errorf("cannot get reference sequence object: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, "", b.operationError("cannot get reference sequence object")
	}

	rawType, err := b.output()
	if err != nil {
		return 0, "", fmt.Errorf("cannot copy reference sequence object type: %w", err)
	}

	return Object(handle), string(rawType), nil
}

func (b *Engine) DeleteMap(
	object Object,
	key string,
) error {
	pointer, length, err := b.write([]byte(key))
	if err != nil {
		return fmt.Errorf("cannot write deleted map key: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_delete_map",
		uint64(object),
		uint64(pointer),
		uint64(length),
	); err != nil {
		return fmt.Errorf("cannot delete reference map value: %w", err)
	}

	return nil
}

func (b *Engine) DeleteSequence(
	object Object,
	index uint64,
) error {
	if err := b.run(
		"am_delete_sequence",
		uint64(object),
		index,
	); err != nil {
		return fmt.Errorf("cannot delete reference sequence value: %w", err)
	}

	return nil
}

func (b *Engine) Increment(
	object Object,
	key string,
	delta int64,
) error {
	pointer, length, err := b.write([]byte(key))
	if err != nil {
		return fmt.Errorf("cannot write incremented map key: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_increment",
		uint64(object),
		uint64(pointer),
		uint64(length),
		uint64(delta),
	); err != nil {
		return fmt.Errorf("cannot increment reference counter: %w", err)
	}

	return nil
}

func (b *Engine) IncrementAt(
	object Object,
	index uint64,
	delta int64,
) error {
	if err := b.run(
		"am_increment_at",
		uint64(object),
		index,
		uint64(delta),
	); err != nil {
		return fmt.Errorf("cannot increment reference sequence counter: %w", err)
	}

	return nil
}

func (b *Engine) Keys(
	object Object,
) ([]string, error) {
	if err := b.run("am_keys", uint64(object)); err != nil {
		return nil, fmt.Errorf("cannot get reference map keys: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference map keys: %w", err)
	}

	var keys []string
	if err := json.Unmarshal(output, &keys); err != nil {
		return nil, fmt.Errorf("cannot decode reference map keys: %w", err)
	}

	return keys, nil
}

func (b *Engine) Length(
	object Object,
) (uint64, error) {
	result, err := b.call("am_length", uint64(object))
	if err != nil {
		return 0, fmt.Errorf("cannot get reference object length: %w", err)
	}

	length := int64(result[0])
	if length < 0 {
		return 0, b.operationError("cannot get reference object length")
	}

	return uint64(length), nil
}

func (b *Engine) PutText(object Object, key string) (Object, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return 0, fmt.Errorf("cannot write text key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	result, err := b.call(
		"am_put_text",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot create reference text: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot create reference text")
	}

	return Object(handle), nil
}

func (b *Engine) GetText(object Object, key string) (Object, error) {
	keyPointer, keyLength, err := b.write([]byte(key))
	if err != nil {
		return 0, fmt.Errorf("cannot write text key: %w", err)
	}
	defer b.free(keyPointer, keyLength)

	result, err := b.call(
		"am_get_text",
		uint64(object),
		uint64(keyPointer),
		uint64(keyLength),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot get reference text: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot get reference text")
	}

	return Object(handle), nil
}

func (b *Engine) SpliceText(
	object Object,
	index uint32,
	deleteCount int32,
	value string,
) error {
	valuePointer, valueLength, err := b.write([]byte(value))
	if err != nil {
		return fmt.Errorf("cannot write splice value: %w", err)
	}
	defer b.free(valuePointer, valueLength)

	if err := b.run(
		"am_text_splice",
		uint64(object),
		uint64(index),
		uint64(uint32(deleteCount)),
		uint64(valuePointer),
		uint64(valueLength),
	); err != nil {
		return fmt.Errorf("cannot splice reference text: %w", err)
	}

	return nil
}

func (b *Engine) UpdateText(
	object Object,
	value string,
) error {
	valuePointer, valueLength, err := b.write([]byte(value))
	if err != nil {
		return fmt.Errorf("cannot write update value: %w", err)
	}
	defer b.free(valuePointer, valueLength)

	if err := b.run(
		"am_text_update",
		uint64(object),
		uint64(valuePointer),
		uint64(valueLength),
	); err != nil {
		return fmt.Errorf("cannot update reference text: %w", err)
	}

	return nil
}

func (b *Engine) UpdateSpans(
	object Object,
	spans []byte,
	config []byte,
) error {
	spansPointer, spansLength, err := b.write(spans)
	if err != nil {
		return fmt.Errorf("cannot write update spans: %w", err)
	}
	defer b.free(spansPointer, spansLength)

	configPointer, configLength, err := b.write(config)
	if err != nil {
		return fmt.Errorf("cannot write update spans config: %w", err)
	}
	defer b.free(configPointer, configLength)

	if err := b.run(
		"am_update_spans",
		uint64(object),
		uint64(spansPointer),
		uint64(spansLength),
		uint64(configPointer),
		uint64(configLength),
	); err != nil {
		return fmt.Errorf("cannot update reference spans: %w", err)
	}

	return nil
}

func (b *Engine) MarkText(
	object Object,
	start uint32,
	end uint32,
	name string,
	value []byte,
	expand string,
) error {
	namePointer, nameLength, err := b.write([]byte(name))
	if err != nil {
		return fmt.Errorf("cannot write mark name: %w", err)
	}
	defer b.free(namePointer, nameLength)

	valuePointer, valueLength, err := b.write(value)
	if err != nil {
		return fmt.Errorf("cannot write mark value: %w", err)
	}
	defer b.free(valuePointer, valueLength)

	expandPointer, expandLength, err := b.write([]byte(expand))
	if err != nil {
		return fmt.Errorf("cannot write mark expansion: %w", err)
	}
	defer b.free(expandPointer, expandLength)

	if err := b.run(
		"am_text_mark",
		uint64(object),
		uint64(start),
		uint64(end),
		uint64(namePointer),
		uint64(nameLength),
		uint64(valuePointer),
		uint64(valueLength),
		uint64(expandPointer),
		uint64(expandLength),
	); err != nil {
		return fmt.Errorf("cannot mark reference text: %w", err)
	}

	return nil
}

func (b *Engine) SplitBlock(
	object Object,
	index uint32,
) (Object, error) {
	result, err := b.call(
		"am_split_block",
		uint64(object),
		uint64(index),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot split reference block: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot split reference block")
	}

	return Object(handle), nil
}

func (b *Engine) JoinBlock(
	object Object,
	index uint32,
) error {
	if err := b.run(
		"am_join_block",
		uint64(object),
		uint64(index),
	); err != nil {
		return fmt.Errorf("cannot join reference block: %w", err)
	}

	return nil
}

func (b *Engine) ReplaceBlock(
	object Object,
	index uint32,
) (Object, error) {
	result, err := b.call(
		"am_replace_block",
		uint64(object),
		uint64(index),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot replace reference block: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot replace reference block")
	}

	return Object(handle), nil
}

func (b *Engine) Text(object Object) (string, error) {
	if err := b.run("am_text", uint64(object)); err != nil {
		return "", fmt.Errorf("cannot read reference text: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return "", fmt.Errorf("cannot copy reference text: %w", err)
	}

	return string(output), nil
}

func (b *Engine) TextAt(
	object Object,
	heads [][32]byte,
) (string, error) {
	pointer, length, err := b.write(flattenHashes(heads))
	if err != nil {
		return "", fmt.Errorf("cannot write historical text heads: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_text_at",
		uint64(object),
		uint64(pointer),
		uint64(length),
	); err != nil {
		return "", fmt.Errorf("cannot read historical reference text: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return "", fmt.Errorf("cannot copy historical reference text: %w", err)
	}

	return string(output), nil
}

func (b *Engine) TextSpans(object Object) ([]byte, error) {
	if err := b.run("am_text_spans", uint64(object)); err != nil {
		return nil, fmt.Errorf("cannot read reference text spans: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference text spans: %w", err)
	}

	return output, nil
}

func (b *Engine) TextSpansAt(
	object Object,
	heads [][32]byte,
) ([]byte, error) {
	pointer, length, err := b.write(flattenHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot write historical span heads: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_text_spans_at",
		uint64(object),
		uint64(pointer),
		uint64(length),
	); err != nil {
		return nil, fmt.Errorf("cannot read historical reference text spans: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy historical reference text spans: %w", err)
	}

	return output, nil
}

func (b *Engine) Marks(object Object) ([]byte, error) {
	if err := b.run("am_marks", uint64(object)); err != nil {
		return nil, fmt.Errorf("cannot read reference marks: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference marks: %w", err)
	}

	return output, nil
}

func (b *Engine) MarksAt(
	object Object,
	heads [][32]byte,
) ([]byte, error) {
	pointer, length, err := b.write(flattenHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot write historical mark heads: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_marks_at",
		uint64(object),
		uint64(pointer),
		uint64(length),
	); err != nil {
		return nil, fmt.Errorf("cannot read historical reference marks: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy historical reference marks: %w", err)
	}

	return output, nil
}

func (b *Engine) TextCursor(object Object, index uint32) ([]byte, error) {
	if err := b.run(
		"am_text_cursor",
		uint64(object),
		uint64(index),
	); err != nil {
		return nil, fmt.Errorf("cannot create reference text cursor: %w", err)
	}

	cursor, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference text cursor: %w", err)
	}

	return cursor, nil
}

func (b *Engine) TextCursorMoving(
	object Object,
	index uint32,
	moveBefore bool,
) ([]byte, error) {
	var movement uint64
	if moveBefore {
		movement = 1
	}

	if err := b.run(
		"am_text_cursor_moving",
		uint64(object),
		uint64(index),
		movement,
	); err != nil {
		return nil, fmt.Errorf("cannot create moving reference text cursor: %w", err)
	}

	cursor, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy moving reference text cursor: %w", err)
	}

	return cursor, nil
}

func (b *Engine) TextCursorMovingAt(
	object Object,
	index uint32,
	moveBefore bool,
	heads [][32]byte,
) ([]byte, error) {
	var movement uint64
	if moveBefore {
		movement = 1
	}

	pointer, length, err := b.write(flattenHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot write cursor heads: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_text_cursor_moving_at",
		uint64(object),
		uint64(index),
		movement,
		uint64(pointer),
		uint64(length),
	); err != nil {
		return nil, fmt.Errorf("cannot create historical reference text cursor: %w", err)
	}

	cursor, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy historical reference text cursor: %w", err)
	}

	return cursor, nil
}

func (b *Engine) TextCursorPosition(
	object Object,
	cursor []byte,
) (uint32, error) {
	pointer, length, err := b.write(cursor)
	if err != nil {
		return 0, fmt.Errorf("cannot write reference text cursor: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call(
		"am_text_cursor_position",
		uint64(object),
		uint64(pointer),
		uint64(length),
	)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve reference text cursor: %w", err)
	}

	position := int64(result[0])
	if position < 0 {
		return 0, b.operationError("cannot resolve reference text cursor")
	}

	if position > int64(^uint32(0)) {
		return 0, fmt.Errorf("cannot resolve reference text cursor: position %d exceeds uint32", position)
	}

	return uint32(position), nil
}

func (b *Engine) Commit(
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	var hash [32]byte

	timestampSeconds := timestamp.Unix()
	if timestamp.IsZero() {
		timestampSeconds = 0
	}

	messagePointer, messageLength, err := b.write([]byte(message))
	if err != nil {
		return hash, fmt.Errorf("cannot write commit message: %w", err)
	}
	defer b.free(messagePointer, messageLength)

	if err := b.run(
		"am_commit",
		uint64(messagePointer),
		uint64(messageLength),
		uint64(timestampSeconds),
	); err != nil {
		return hash, fmt.Errorf("cannot commit reference document: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return hash, fmt.Errorf("cannot copy reference commit hash: %w", err)
	}

	if len(output) != len(hash) {
		return hash, fmt.Errorf("cannot decode reference commit hash: expected 32 bytes, got %d", len(output))
	}

	copy(hash[:], output)

	return hash, nil
}

func (b *Engine) EmptyCommit(
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	var hash [32]byte

	timestampSeconds := timestamp.Unix()
	if timestamp.IsZero() {
		timestampSeconds = 0
	}

	messagePointer, messageLength, err := b.write([]byte(message))
	if err != nil {
		return hash, fmt.Errorf("cannot write empty commit message: %w", err)
	}
	defer b.free(messagePointer, messageLength)

	if err := b.run(
		"am_empty_commit",
		uint64(messagePointer),
		uint64(messageLength),
		uint64(timestampSeconds),
	); err != nil {
		return hash, fmt.Errorf("cannot commit empty reference change: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return hash, fmt.Errorf("cannot copy empty reference change hash: %w", err)
	}

	if len(output) != len(hash) {
		return hash, fmt.Errorf("invalid empty reference change hash length %d", len(output))
	}

	copy(hash[:], output)

	return hash, nil
}

func (b *Engine) Rollback() (uint64, error) {
	result, err := b.call("am_rollback")
	if err != nil {
		return 0, fmt.Errorf("cannot roll back reference document: %w", err)
	}

	cancelled := int64(result[0])
	if cancelled < 0 {
		return 0, b.operationError("cannot roll back reference document")
	}

	return uint64(cancelled), nil
}

func (b *Engine) Stats() ([]byte, error) {
	if err := b.run("am_stats"); err != nil {
		return nil, fmt.Errorf("cannot read reference stats: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference stats: %w", err)
	}

	return output, nil
}

func (b *Engine) CurrentState() ([]byte, error) {
	if err := b.run("am_current_state"); err != nil {
		return nil, fmt.Errorf("cannot read reference current state: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference current state: %w", err)
	}

	return output, nil
}

func (b *Engine) UpdateDiffCursor() error {
	if err := b.run("am_update_diff_cursor"); err != nil {
		return fmt.Errorf("cannot update reference diff cursor: %w", err)
	}

	return nil
}

func (b *Engine) DiffIncremental() ([]byte, error) {
	if err := b.run("am_diff_incremental"); err != nil {
		return nil, fmt.Errorf("cannot read reference incremental diff: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference incremental diff: %w", err)
	}

	return output, nil
}

func (b *Engine) Diff(
	before [][32]byte,
	after [][32]byte,
) ([]byte, error) {
	beforePointer, beforeLength, err := b.write(flattenHashes(before))
	if err != nil {
		return nil, fmt.Errorf("cannot write diff before heads: %w", err)
	}
	defer b.free(beforePointer, beforeLength)

	afterPointer, afterLength, err := b.write(flattenHashes(after))
	if err != nil {
		return nil, fmt.Errorf("cannot write diff after heads: %w", err)
	}
	defer b.free(afterPointer, afterLength)

	if err := b.run(
		"am_diff",
		uint64(beforePointer),
		uint64(beforeLength),
		uint64(afterPointer),
		uint64(afterLength),
	); err != nil {
		return nil, fmt.Errorf("cannot read reference diff: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference diff: %w", err)
	}

	return output, nil
}

func (b *Engine) Heads() ([][32]byte, error) {
	if err := b.run("am_heads"); err != nil {
		return nil, fmt.Errorf("cannot read reference heads: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference heads: %w", err)
	}

	if len(output)%32 != 0 {
		return nil, fmt.Errorf("cannot decode reference heads: length %d is not divisible by 32", len(output))
	}

	heads := make([][32]byte, len(output)/32)
	for i := range heads {
		copy(heads[i][:], output[i*32:(i+1)*32])
	}

	return heads, nil
}

func (b *Engine) HasHeads(
	heads [][32]byte,
) (bool, error) {
	pointer, length, err := b.write(flattenHashes(heads))
	if err != nil {
		return false, fmt.Errorf("cannot write reference heads: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call(
		"am_has_heads",
		uint64(pointer),
		uint64(length),
	)
	if err != nil {
		return false, fmt.Errorf("cannot inspect reference heads: %w", err)
	}

	value := int32(result[0])
	if value < 0 {
		return false, b.operationError("cannot inspect reference heads")
	}

	return value != 0, nil
}

// BloomContains builds a sync Bloom filter from the seed change hashes and
// reports whether it contains the target hash. Because Bloom filters admit
// false positives, a true result does not guarantee membership; parity tests
// use this to reproduce the upstream false-positive search deterministically.
func (b *Engine) BloomContains(
	target [32]byte,
	seeds [][32]byte,
) (bool, error) {
	input := make([]byte, 0, (len(seeds)+1)*32)
	input = append(input, target[:]...)
	input = append(input, flattenHashes(seeds)...)

	pointer, length, err := b.write(input)
	if err != nil {
		return false, fmt.Errorf("cannot write bloom hashes: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call(
		"am_bloom_contains",
		uint64(pointer),
		uint64(length),
	)
	if err != nil {
		return false, fmt.Errorf("cannot evaluate reference bloom filter: %w", err)
	}

	value := int32(result[0])
	if value < 0 {
		return false, b.operationError("cannot evaluate reference bloom filter")
	}

	return value != 0, nil
}

func (b *Engine) MissingDependencies(
	heads [][32]byte,
) ([][32]byte, error) {
	pointer, length, err := b.write(flattenHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot write dependency heads: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_missing_dependencies",
		uint64(pointer),
		uint64(length),
	); err != nil {
		return nil, fmt.Errorf("cannot get reference missing dependencies: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference missing dependencies: %w", err)
	}

	if len(output)%32 != 0 {
		return nil, fmt.Errorf(
			"invalid reference dependency byte length %d",
			len(output),
		)
	}

	result := make([][32]byte, len(output)/32)
	for i := range result {
		copy(result[i][:], output[i*32:(i+1)*32])
	}

	return result, nil
}

func (b *Engine) Merge(other []byte) ([][32]byte, error) {
	if err := b.runBytes("am_merge", other); err != nil {
		return nil, fmt.Errorf("cannot merge reference document: %w", err)
	}

	output, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy merged reference heads: %w", err)
	}

	if len(output)%32 != 0 {
		return nil, fmt.Errorf("cannot decode merged reference heads: length %d is not divisible by 32", len(output))
	}

	heads := make([][32]byte, len(output)/32)
	for i := range heads {
		copy(heads[i][:], output[i*32:(i+1)*32])
	}

	return heads, nil
}

func (b *Engine) NewSyncState() (uint32, error) {
	result, err := b.call("am_sync_new")
	if err != nil {
		return 0, fmt.Errorf("cannot create reference sync state: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot create reference sync state")
	}

	return uint32(handle), nil
}

func (b *Engine) CloseSyncState(handle uint32) error {
	if err := b.run("am_sync_free", uint64(handle)); err != nil {
		return fmt.Errorf("cannot close reference sync state: %w", err)
	}

	return nil
}

func (b *Engine) SetSyncReadOnly(
	handle uint32,
	readOnly bool,
) error {
	var value uint64
	if readOnly {
		value = 1
	}

	if err := b.run(
		"am_sync_set_read_only",
		uint64(handle),
		value,
	); err != nil {
		return fmt.Errorf("cannot set reference sync read-only mode: %w", err)
	}

	return nil
}

func (b *Engine) SyncPeerReadOnly(
	handle uint32,
) (bool, error) {
	result, err := b.call(
		"am_sync_peer_read_only",
		uint64(handle),
	)
	if err != nil {
		return false, fmt.Errorf("cannot get reference peer read-only mode: %w", err)
	}

	value := int32(result[0])
	if value < 0 {
		return false, b.operationError("cannot get reference peer read-only mode")
	}

	return value != 0, nil
}

func (b *Engine) GenerateSyncMessage(
	handle uint32,
) ([]byte, bool, error) {
	if err := b.run("am_sync_generate", uint64(handle)); err != nil {
		return nil, false, fmt.Errorf("cannot generate reference sync message: %w", err)
	}

	message, err := b.output()
	if err != nil {
		return nil, false, fmt.Errorf("cannot copy reference sync message: %w", err)
	}

	return message, len(message) > 0, nil
}

func (b *Engine) ReceiveSyncMessage(handle uint32, message []byte) error {
	pointer, length, err := b.write(message)
	if err != nil {
		return fmt.Errorf("cannot write reference sync message: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(
		"am_sync_receive",
		uint64(handle),
		uint64(pointer),
		uint64(length),
	); err != nil {
		return fmt.Errorf("cannot receive reference sync message: %w", err)
	}

	return nil
}

func (b *Engine) SaveSyncState(handle uint32) ([]byte, error) {
	if err := b.run("am_sync_save", uint64(handle)); err != nil {
		return nil, fmt.Errorf("cannot save reference sync state: %w", err)
	}

	data, err := b.output()
	if err != nil {
		return nil, fmt.Errorf("cannot copy reference sync state: %w", err)
	}

	return data, nil
}

func (b *Engine) LoadSyncState(data []byte) (uint32, error) {
	pointer, length, err := b.write(data)
	if err != nil {
		return 0, fmt.Errorf("cannot write reference sync state: %w", err)
	}
	defer b.free(pointer, length)

	result, err := b.call("am_sync_load", uint64(pointer), uint64(length))
	if err != nil {
		return 0, fmt.Errorf("cannot load reference sync state: %w", err)
	}

	handle := int64(result[0])
	if handle < 0 {
		return 0, b.operationError("cannot load reference sync state")
	}

	return uint32(handle), nil
}

func (b *Engine) runBytes(function string, value []byte) error {
	pointer, length, err := b.write(value)
	if err != nil {
		return fmt.Errorf("cannot write operation input: %w", err)
	}
	defer b.free(pointer, length)

	if err := b.run(function, uint64(pointer), uint64(length)); err != nil {
		return fmt.Errorf("cannot run operation with input: %w", err)
	}

	return nil
}

func (b *Engine) run(function string, parameters ...uint64) error {
	result, err := b.call(function, parameters...)
	if err != nil {
		return fmt.Errorf("cannot call %s: %w", function, err)
	}

	if int32(result[0]) != 0 {
		return b.operationError(fmt.Sprintf("%s failed", function))
	}

	return nil
}

func (b *Engine) operationError(fallback string) error {
	result, err := b.call("am_error_len")
	if err != nil {
		return fmt.Errorf("%s: cannot read error length: %w", fallback, err)
	}

	length := uint32(result[0])
	if length == 0 {
		return fmt.Errorf("%s", fallback)
	}

	pointer, err := b.alloc(length)
	if err != nil {
		return fmt.Errorf("%s: cannot allocate error buffer: %w", fallback, err)
	}
	defer b.free(pointer, length)

	if _, err := b.call("am_error_copy", uint64(pointer)); err != nil {
		return fmt.Errorf("%s: cannot copy error: %w", fallback, err)
	}

	message, err := b.read(pointer, length)
	if err != nil {
		return fmt.Errorf("%s: cannot read error: %w", fallback, err)
	}

	return fmt.Errorf("%s: %s", fallback, message)
}

func (b *Engine) output() ([]byte, error) {
	result, err := b.call("am_output_len")
	if err != nil {
		return nil, fmt.Errorf("cannot read output length: %w", err)
	}

	length := uint32(result[0])
	if length == 0 {
		return nil, nil
	}

	pointer, err := b.alloc(length)
	if err != nil {
		return nil, fmt.Errorf("cannot allocate output buffer: %w", err)
	}
	defer b.free(pointer, length)

	if _, err := b.call("am_output_copy", uint64(pointer)); err != nil {
		return nil, fmt.Errorf("cannot copy output: %w", err)
	}

	output, err := b.read(pointer, length)
	if err != nil {
		return nil, fmt.Errorf("cannot read output: %w", err)
	}

	return output, nil
}

func (b *Engine) call(function string, parameters ...uint64) ([]uint64, error) {
	exported := b.module.ExportedFunction(function)
	if exported == nil {
		return nil, fmt.Errorf("reference module does not export %q", function)
	}

	result, err := exported.Call(context.Background(), parameters...)
	if err != nil {
		return nil, fmt.Errorf("cannot execute reference export %q: %w", function, err)
	}

	return result, nil
}

func (b *Engine) alloc(length uint32) (uint32, error) {
	result, err := b.call("am_alloc", uint64(length))
	if err != nil {
		return 0, fmt.Errorf("cannot allocate reference memory: %w", err)
	}

	pointer := uint32(result[0])
	if pointer == 0 && length > 0 {
		return 0, fmt.Errorf("cannot allocate %d bytes of reference memory", length)
	}

	return pointer, nil
}

func (b *Engine) free(pointer, length uint32) {
	if pointer == 0 || length == 0 {
		return
	}

	_, _ = b.call("am_free", uint64(pointer), uint64(length))
}

func (b *Engine) write(value []byte) (uint32, uint32, error) {
	length := uint32(len(value))
	if length == 0 {
		return 0, 0, nil
	}

	pointer, err := b.alloc(length)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot allocate input: %w", err)
	}

	if !b.module.Memory().Write(pointer, value) {
		b.free(pointer, length)
		return 0, 0, fmt.Errorf("cannot write %d bytes at reference memory offset %d", length, pointer)
	}

	return pointer, length, nil
}

func (b *Engine) read(pointer, length uint32) ([]byte, error) {
	value, ok := b.module.Memory().Read(pointer, length)
	if !ok {
		return nil, fmt.Errorf("cannot read %d bytes at reference memory offset %d", length, pointer)
	}

	return append([]byte(nil), value...), nil
}

func flattenHashes(hashes [][32]byte) []byte {
	value := make([]byte, 0, len(hashes)*32)
	for _, hash := range hashes {
		value = append(value, hash[:]...)
	}

	return value
}
