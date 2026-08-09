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

use std::cell::RefCell;
use std::mem;
use std::ptr;
use std::slice;

use automerge::marks::{ExpandMark, Mark};
use automerge::sync::{Message, State as SyncState, SyncDoc};
use automerge::transaction::{CommitOptions, Transactable};
use automerge::{
    ActorId, AutoCommit, ChangeHash, Cursor, MoveCursor, ObjId, ObjType, Patch, PatchAction, Prop,
    ReadDoc, ScalarValue, Value, ROOT,
};

struct State {
    doc: AutoCommit,
    objects: Vec<ObjId>,
    sync_states: Vec<Option<SyncState>>,
    output: Vec<u8>,
    error: String,
}

impl State {
    fn new() -> Self {
        Self {
            doc: AutoCommit::new(),
            objects: vec![ROOT],
            sync_states: Vec::new(),
            output: Vec::new(),
            error: String::new(),
        }
    }

    fn reset(&mut self, doc: AutoCommit) {
        self.doc = doc;
        self.objects.clear();
        self.objects.push(ROOT);
        self.sync_states.clear();
        self.output.clear();
        self.error.clear();
    }

    fn object(&self, handle: u32) -> Result<ObjId, String> {
        self.objects
            .get(handle as usize)
            .cloned()
            .ok_or_else(|| format!("invalid object handle {handle}"))
    }

    fn push_object(&mut self, object: ObjId) -> Result<u32, String> {
        let handle =
            u32::try_from(self.objects.len()).map_err(|_| "too many object handles".to_owned())?;
        self.objects.push(object);
        Ok(handle)
    }

    fn fail(&mut self, error: impl ToString) -> i32 {
        self.error = error.to_string();
        -1
    }
}

thread_local! {
    static STATE: RefCell<State> = RefCell::new(State::new());
}

#[no_mangle]
pub extern "C" fn am_abi_version() -> u32 {
    1
}

fn input_bytes(pointer: u32, length: u32) -> Vec<u8> {
    if length == 0 {
        return Vec::new();
    }

    // The caller allocates this memory through am_alloc and keeps it alive for
    // the duration of the call.
    unsafe { slice::from_raw_parts(pointer as *const u8, length as usize).to_vec() }
}

fn input_string(pointer: u32, length: u32) -> Result<String, String> {
    String::from_utf8(input_bytes(pointer, length)).map_err(|error| error.to_string())
}

fn input_scalar(pointer: u32, length: u32) -> Result<ScalarValue, String> {
    let value: serde_json::Value =
        serde_json::from_slice(&input_bytes(pointer, length)).map_err(|error| error.to_string())?;
    let scalar_type = value
        .get("type")
        .and_then(serde_json::Value::as_str)
        .ok_or_else(|| "scalar type is missing".to_owned())?;

    match scalar_type {
        "null" => Ok(ScalarValue::Null),
        "boolean" => value
            .get("bool")
            .and_then(serde_json::Value::as_bool)
            .map(ScalarValue::Boolean)
            .ok_or_else(|| "scalar boolean is missing".to_owned()),
        "uint" => value
            .get("uint")
            .and_then(serde_json::Value::as_u64)
            .map(ScalarValue::Uint)
            .ok_or_else(|| "scalar uint is missing".to_owned()),
        "int" => value
            .get("int")
            .and_then(serde_json::Value::as_i64)
            .map(ScalarValue::Int)
            .ok_or_else(|| "scalar int is missing".to_owned()),
        "float64" => value
            .get("floatBits")
            .and_then(serde_json::Value::as_u64)
            .map(|bits| ScalarValue::F64(f64::from_bits(bits)))
            .ok_or_else(|| "scalar float bits are missing".to_owned()),
        "string" => value
            .get("string")
            .and_then(serde_json::Value::as_str)
            .map(|value| ScalarValue::Str(value.into()))
            .ok_or_else(|| "scalar string is missing".to_owned()),
        "bytes" => value
            .get("bytes")
            .and_then(serde_json::Value::as_str)
            .ok_or_else(|| "scalar bytes are missing".to_owned())
            .and_then(|value| hex::decode(value).map_err(|error| error.to_string()))
            .map(ScalarValue::Bytes),
        "counter" => value
            .get("int")
            .and_then(serde_json::Value::as_i64)
            .map(|value| ScalarValue::Counter(value.into()))
            .ok_or_else(|| "scalar counter is missing".to_owned()),
        "timestamp" => value
            .get("int")
            .and_then(serde_json::Value::as_i64)
            .map(ScalarValue::Timestamp)
            .ok_or_else(|| "scalar timestamp is missing".to_owned()),
        other => Err(format!("unknown scalar type {other:?}")),
    }
}

fn input_heads(pointer: u32, length: u32) -> Result<Vec<ChangeHash>, String> {
    let bytes = input_bytes(pointer, length);
    if bytes.len() % 32 != 0 {
        return Err("head bytes are not a multiple of 32".to_owned());
    }

    bytes
        .chunks_exact(32)
        .map(|value| {
            let mut hash = [0_u8; 32];
            hash.copy_from_slice(value);
            Ok(ChangeHash(hash))
        })
        .collect()
}

fn scalar_json(value: &ScalarValue) -> Result<serde_json::Value, String> {
    let value = match value {
        ScalarValue::Null => serde_json::json!({"type": "null"}),
        ScalarValue::Boolean(value) => {
            serde_json::json!({"type": "boolean", "bool": value})
        }
        ScalarValue::Uint(value) => serde_json::json!({"type": "uint", "uint": value}),
        ScalarValue::Int(value) => serde_json::json!({"type": "int", "int": value}),
        ScalarValue::F64(value) => {
            serde_json::json!({"type": "float64", "floatBits": value.to_bits()})
        }
        ScalarValue::Str(value) => serde_json::json!({"type": "string", "string": value}),
        ScalarValue::Bytes(value) => {
            serde_json::json!({"type": "bytes", "bytes": hex::encode(value)})
        }
        ScalarValue::Counter(value) => {
            serde_json::json!({"type": "counter", "int": i64::from(value)})
        }
        ScalarValue::Timestamp(value) => {
            serde_json::json!({"type": "timestamp", "int": value})
        }
        ScalarValue::Unknown { .. } => return Err("unknown scalar type is unsupported".to_owned()),
    };

    Ok(value)
}

fn encode_scalar(value: &ScalarValue) -> Result<Vec<u8>, String> {
    serde_json::to_vec(&scalar_json(value)?).map_err(|error| error.to_string())
}

fn parse_object_type(value: &str) -> Result<ObjType, String> {
    match value {
        "map" => Ok(ObjType::Map),
        "list" => Ok(ObjType::List),
        "text" => Ok(ObjType::Text),
        "table" => Ok(ObjType::Table),
        other => Err(format!("unknown object type {other:?}")),
    }
}

fn encode_object_type(value: ObjType) -> &'static str {
    match value {
        ObjType::Map => "map",
        ObjType::List => "list",
        ObjType::Text => "text",
        ObjType::Table => "table",
    }
}

fn parse_mark_expand(value: &str) -> Result<ExpandMark, String> {
    match value {
        "before" => Ok(ExpandMark::Before),
        "after" => Ok(ExpandMark::After),
        "both" => Ok(ExpandMark::Both),
        "none" => Ok(ExpandMark::None),
        other => Err(format!("unknown mark expansion {other:?}")),
    }
}

#[no_mangle]
pub extern "C" fn am_alloc(length: u32) -> u32 {
    if length == 0 {
        return 0;
    }

    let mut bytes = Vec::<u8>::with_capacity(length as usize);
    let pointer = bytes.as_mut_ptr();
    mem::forget(bytes);
    pointer as u32
}

#[no_mangle]
pub extern "C" fn am_free(pointer: u32, length: u32) {
    if pointer == 0 || length == 0 {
        return;
    }

    unsafe {
        drop(Vec::from_raw_parts(pointer as *mut u8, 0, length as usize));
    }
}

#[no_mangle]
pub extern "C" fn am_output_len() -> u32 {
    STATE.with(|state| state.borrow().output.len() as u32)
}

#[no_mangle]
pub extern "C" fn am_output_copy(pointer: u32) -> i32 {
    STATE.with(|state| {
        let state = state.borrow();
        if !state.output.is_empty() && pointer == 0 {
            return -1;
        }

        unsafe {
            ptr::copy_nonoverlapping(
                state.output.as_ptr(),
                pointer as *mut u8,
                state.output.len(),
            );
        }
        0
    })
}

#[no_mangle]
pub extern "C" fn am_error_len() -> u32 {
    STATE.with(|state| state.borrow().error.len() as u32)
}

#[no_mangle]
pub extern "C" fn am_error_copy(pointer: u32) -> i32 {
    STATE.with(|state| {
        let state = state.borrow();
        if !state.error.is_empty() && pointer == 0 {
            return -1;
        }

        unsafe {
            ptr::copy_nonoverlapping(state.error.as_ptr(), pointer as *mut u8, state.error.len());
        }
        0
    })
}

#[no_mangle]
pub extern "C" fn am_create() -> i32 {
    STATE.with(|state| state.borrow_mut().reset(AutoCommit::new()));
    0
}

#[no_mangle]
pub extern "C" fn am_load(pointer: u32, length: u32) -> i32 {
    let bytes = input_bytes(pointer, length);
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        match AutoCommit::load(&bytes) {
            Ok(doc) => {
                state.reset(doc);
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_save() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        state.output = state.doc.save();
        state.error.clear();
    });
    0
}

#[no_mangle]
pub extern "C" fn am_save_incremental() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        state.output = state.doc.save_incremental();
        state.error.clear();
    });
    0
}

#[no_mangle]
pub extern "C" fn am_load_incremental(pointer: u32, length: u32) -> i64 {
    let bytes = input_bytes(pointer, length);
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        match state.doc.load_incremental(&bytes) {
            Ok(applied) => match i64::try_from(applied) {
                Ok(applied) => {
                    state.error.clear();
                    applied
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_set_actor(pointer: u32, length: u32) -> i32 {
    let actor = input_bytes(pointer, length);
    if actor.is_empty() {
        return STATE.with(|state| state.borrow_mut().fail("actor ID cannot be empty"));
    }

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        state.doc.set_actor(ActorId::from(actor));
        state.error.clear();
    });
    0
}

#[no_mangle]
pub extern "C" fn am_put_string(
    object_handle: u32,
    key_pointer: u32,
    key_length: u32,
    value_pointer: u32,
    value_length: u32,
) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    let value = match input_string(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.put(&object, key, value) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_string(object_handle: u32, key_pointer: u32, key_length: u32) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.get(&object, key) {
            Ok(Some((Value::Scalar(value), _))) => match value.as_ref() {
                ScalarValue::Str(value) => {
                    state.output = value.as_bytes().to_vec();
                    state.error.clear();
                    0
                }
                _ => state.fail("value is not a string"),
            },
            Ok(Some(_)) => state.fail("value is not a scalar"),
            Ok(None) => state.fail("value does not exist"),
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_put_scalar(
    object_handle: u32,
    key_pointer: u32,
    key_length: u32,
    value_pointer: u32,
    value_length: u32,
) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    let value = match input_scalar(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.put(&object, key, value) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_scalar(object_handle: u32, key_pointer: u32, key_length: u32) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.get(&object, key) {
            Ok(Some((Value::Scalar(value), _))) => match encode_scalar(value.as_ref()) {
                Ok(encoded) => {
                    state.output = encoded;
                    state.error.clear();
                    0
                }
                Err(error) => state.fail(error),
            },
            Ok(Some(_)) => state.fail("value is not a scalar"),
            Ok(None) => state.fail("value does not exist"),
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_scalar_at_heads(
    object_handle: u32,
    key_pointer: u32,
    key_length: u32,
    heads_pointer: u32,
    heads_length: u32,
) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    let heads = match input_heads(heads_pointer, heads_length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.get_at(&object, key, &heads) {
            Ok(Some((Value::Scalar(value), _))) => match encode_scalar(value.as_ref()) {
                Ok(encoded) => {
                    state.output = encoded;
                    state.error.clear();
                    0
                }
                Err(error) => state.fail(error),
            },
            Ok(Some(_)) => state.fail("historical value is not a scalar"),
            Ok(None) => state.fail("historical value does not exist"),
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_all_scalars(object_handle: u32, key_pointer: u32, key_length: u32) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.get_all(&object, key) {
            Ok(values) => {
                if values.is_empty() {
                    return state.fail("scalar property does not exist");
                }
                let encoded = values
                    .iter()
                    .filter_map(|(value, _)| match value {
                        Value::Scalar(value) => Some(scalar_json(value.as_ref())),
                        Value::Object(_) => None,
                    })
                    .collect::<Result<Vec<_>, _>>();
                match encoded.and_then(|values| {
                    serde_json::to_vec(&values).map_err(|error| error.to_string())
                }) {
                    Ok(encoded) => {
                        state.output = encoded;
                        state.error.clear();
                        0
                    }
                    Err(error) => state.fail(error),
                }
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_all_scalars_at(object_handle: u32, index: u64) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => return state.fail("sequence index exceeds platform capacity"),
        };
        match state.doc.get_all(&object, index) {
            Ok(values) => {
                if values.is_empty() {
                    return state.fail("sequence value does not exist");
                }
                let encoded = values
                    .iter()
                    .filter_map(|(value, _)| match value {
                        Value::Scalar(value) => Some(scalar_json(value.as_ref())),
                        Value::Object(_) => None,
                    })
                    .collect::<Result<Vec<_>, _>>();
                match encoded.and_then(|values| {
                    serde_json::to_vec(&values).map_err(|error| error.to_string())
                }) {
                    Ok(encoded) => {
                        state.output = encoded;
                        state.error.clear();
                        0
                    }
                    Err(error) => state.fail(error),
                }
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_put_object(
    object_handle: u32,
    key_pointer: u32,
    key_length: u32,
    type_pointer: u32,
    type_length: u32,
) -> i64 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => {
            STATE.with(|state| state.borrow_mut().fail(error));
            return -1;
        }
    };
    let object_type =
        match input_string(type_pointer, type_length).and_then(|value| parse_object_type(&value)) {
            Ok(object_type) => object_type,
            Err(error) => {
                STATE.with(|state| state.borrow_mut().fail(error));
                return -1;
            }
        };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.put_object(&object, key, object_type) {
            Ok(object) => match state.push_object(object) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_object(object_handle: u32, key_pointer: u32, key_length: u32) -> i64 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => {
            STATE.with(|state| state.borrow_mut().fail(error));
            return -1;
        }
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.get(&object, key) {
            Ok(Some((Value::Object(object_type), object))) => match state.push_object(object) {
                Ok(handle) => {
                    state.output = encode_object_type(object_type).as_bytes().to_vec();
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Ok(Some(_)) => {
                state.fail("value is not an object");
                -1
            }
            Ok(None) => {
                state.fail("object does not exist");
                -1
            }
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_insert_scalar(
    object_handle: u32,
    index: u64,
    value_pointer: u32,
    value_length: u32,
) -> i32 {
    let value = match input_scalar(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => return state.fail("sequence index exceeds platform capacity"),
        };
        match state.doc.insert(&object, index, value) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_put_scalar_at(
    object_handle: u32,
    index: u64,
    value_pointer: u32,
    value_length: u32,
) -> i32 {
    let value = match input_scalar(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => return state.fail("sequence index exceeds platform capacity"),
        };
        match state.doc.put(&object, index, value) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_insert_object(
    object_handle: u32,
    index: u64,
    type_pointer: u32,
    type_length: u32,
) -> i64 {
    let object_type =
        match input_string(type_pointer, type_length).and_then(|value| parse_object_type(&value)) {
            Ok(object_type) => object_type,
            Err(error) => {
                STATE.with(|state| state.borrow_mut().fail(error));
                return -1;
            }
        };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => {
                state.fail("sequence index exceeds platform capacity");
                return -1;
            }
        };
        match state.doc.insert_object(&object, index, object_type) {
            Ok(object) => match state.push_object(object) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_put_object_at(
    object_handle: u32,
    index: u64,
    type_pointer: u32,
    type_length: u32,
) -> i64 {
    let object_type =
        match input_string(type_pointer, type_length).and_then(|value| parse_object_type(&value)) {
            Ok(object_type) => object_type,
            Err(error) => {
                STATE.with(|state| state.borrow_mut().fail(error));
                return -1;
            }
        };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => {
                state.fail("sequence index exceeds platform capacity");
                return -1;
            }
        };
        match state.doc.put_object(&object, index, object_type) {
            Ok(object) => match state.push_object(object) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_scalar_at(object_handle: u32, index: u64) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => return state.fail("sequence index exceeds platform capacity"),
        };
        match state.doc.get(&object, index) {
            Ok(Some((Value::Scalar(value), _))) => match encode_scalar(value.as_ref()) {
                Ok(encoded) => {
                    state.output = encoded;
                    state.error.clear();
                    0
                }
                Err(error) => state.fail(error),
            },
            Ok(Some(_)) => state.fail("value is not a scalar"),
            Ok(None) => state.fail("sequence value does not exist"),
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_object_at(object_handle: u32, index: u64) -> i64 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => {
                state.fail("sequence index exceeds platform capacity");
                return -1;
            }
        };
        match state.doc.get(&object, index) {
            Ok(Some((Value::Object(object_type), object))) => match state.push_object(object) {
                Ok(handle) => {
                    state.output = encode_object_type(object_type).as_bytes().to_vec();
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Ok(Some(_)) => {
                state.fail("value is not an object");
                -1
            }
            Ok(None) => {
                state.fail("sequence object does not exist");
                -1
            }
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_delete_map(object_handle: u32, key_pointer: u32, key_length: u32) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.delete(&object, key) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_delete_sequence(object_handle: u32, index: u64) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => return state.fail("sequence index exceeds platform capacity"),
        };
        match state.doc.delete(&object, index) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_increment(
    object_handle: u32,
    key_pointer: u32,
    key_length: u32,
    delta: i64,
) -> i32 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.increment(&object, key, delta) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_increment_at(object_handle: u32, index: u64, delta: i64) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let index = match usize::try_from(index) {
            Ok(index) => index,
            Err(_) => return state.fail("sequence index exceeds platform capacity"),
        };
        match state.doc.increment(&object, index, delta) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_keys(object_handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let mut keys = state.doc.keys(&object).collect::<Vec<_>>();
        keys.sort();
        match serde_json::to_vec(&keys) {
            Ok(encoded) => {
                state.output = encoded;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_length(object_handle: u32) -> i64 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match i64::try_from(state.doc.length(&object)) {
            Ok(length) => {
                state.error.clear();
                length
            }
            Err(_) => {
                state.fail("object length exceeds i64");
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_put_text(object_handle: u32, key_pointer: u32, key_length: u32) -> i64 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => {
            STATE.with(|state| state.borrow_mut().fail(error));
            return -1;
        }
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.put_object(&object, key, ObjType::Text) {
            Ok(text) => match state.push_object(text) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_get_text(object_handle: u32, key_pointer: u32, key_length: u32) -> i64 {
    let key = match input_string(key_pointer, key_length) {
        Ok(key) => key,
        Err(error) => {
            STATE.with(|state| state.borrow_mut().fail(error));
            return -1;
        }
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.get(&object, key) {
            Ok(Some((Value::Object(ObjType::Text), text))) => match state.push_object(text) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Ok(Some(_)) => {
                state.fail("value is not text");
                -1
            }
            Ok(None) => {
                state.fail("text does not exist");
                -1
            }
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_splice(
    object_handle: u32,
    index: u32,
    delete_count: i32,
    value_pointer: u32,
    value_length: u32,
) -> i32 {
    let value = match input_string(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state
            .doc
            .splice_text(&object, index as usize, delete_count as isize, &value)
        {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_update(
    object_handle: u32,
    value_pointer: u32,
    value_length: u32,
) -> i32 {
    let value = match input_string(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.update_text(&object, &value) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_mark(
    object_handle: u32,
    start: u32,
    end: u32,
    name_pointer: u32,
    name_length: u32,
    value_pointer: u32,
    value_length: u32,
    expand_pointer: u32,
    expand_length: u32,
) -> i32 {
    let name = match input_string(name_pointer, name_length) {
        Ok(name) => name,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    let value = match input_scalar(value_pointer, value_length) {
        Ok(value) => value,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    let expand = match input_string(expand_pointer, expand_length)
        .and_then(|value| parse_mark_expand(&value))
    {
        Ok(expand) => expand,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let mark = Mark::new(name, value, start as usize, end as usize);
        match state.doc.mark(&object, mark, expand) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_split_block(object_handle: u32, index: u32) -> i64 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.split_block(&object, index as usize) {
            Ok(block) => match state.push_object(block) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_join_block(object_handle: u32, index: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.join_block(&object, index as usize) {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_replace_block(object_handle: u32, index: u32) -> i64 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.replace_block(&object, index as usize) {
            Ok(block) => match state.push_object(block) {
                Ok(handle) => {
                    state.error.clear();
                    i64::from(handle)
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text(object_handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.text(&object) {
            Ok(text) => {
                state.output = text.into_bytes();
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_at(object_handle: u32, heads_pointer: u32, heads_length: u32) -> i32 {
    let heads = match input_heads(heads_pointer, heads_length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.text_at(&object, &heads) {
            Ok(text) => {
                state.output = text.into_bytes();
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_spans(object_handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let spans = match state.doc.spans(&object) {
            Ok(spans) => spans,
            Err(error) => return state.fail(error),
        };

        let values = spans
            .map(|span| match span {
                automerge::Span::Text { text, marks } => {
                    let mut value = serde_json::Map::new();
                    value.insert("type".to_owned(), serde_json::json!("text"));
                    value.insert("value".to_owned(), serde_json::json!(text.to_string()));
                    if let Some(marks) = marks {
                        let marks = marks
                            .iter()
                            .map(|(name, value)| (name.to_string(), scalar_to_json(value)))
                            .collect();
                        value.insert("marks".to_owned(), serde_json::Value::Object(marks));
                    }
                    serde_json::Value::Object(value)
                }
                automerge::Span::Block(block) => serde_json::json!({
                    "type": "block",
                    "value": hydrate_map_to_json(&block),
                }),
            })
            .collect::<Vec<_>>();

        match serde_json::to_vec(&values) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

fn marks_to_output(marks: Vec<automerge::marks::Mark>) -> Result<Vec<u8>, String> {
    let mut values = Vec::with_capacity(marks.len());
    for mark in &marks {
        values.push(serde_json::json!({
            "start": mark.start,
            "end": mark.end,
            "name": mark.name(),
            "value": scalar_json(mark.value())?,
        }));
    }

    serde_json::to_vec(&values).map_err(|error| error.to_string())
}

#[no_mangle]
pub extern "C" fn am_marks(object_handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let marks = match state.doc.marks(&object) {
            Ok(marks) => marks,
            Err(error) => return state.fail(error),
        };
        match marks_to_output(marks) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_marks_at(object_handle: u32, heads_pointer: u32, heads_length: u32) -> i32 {
    let heads = match input_heads(heads_pointer, heads_length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let marks = match state.doc.marks_at(&object, &heads) {
            Ok(marks) => marks,
            Err(error) => return state.fail(error),
        };
        match marks_to_output(marks) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_cursor(object_handle: u32, index: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        match state.doc.get_cursor(&object, index as usize, None) {
            Ok(cursor) => {
                state.output = cursor.to_bytes();
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_text_cursor_moving(object_handle: u32, index: u32, move_before: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => return state.fail(error),
        };
        let movement = if move_before != 0 {
            MoveCursor::Before
        } else {
            MoveCursor::After
        };
        match state
            .doc
            .get_cursor_moving(&object, index as usize, None, movement)
        {
            Ok(cursor) => {
                state.output = cursor.to_bytes();
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

fn hydrate_map_to_json(map: &automerge::hydrate::Map) -> serde_json::Value {
    serde_json::Value::Object(
        map.iter()
            .map(|(key, value)| (key.to_owned(), hydrate_value_to_json(&value.value)))
            .collect(),
    )
}

fn hydrate_value_to_json(value: &automerge::hydrate::Value) -> serde_json::Value {
    match value {
        automerge::hydrate::Value::Scalar(value) => scalar_to_json(value),
        automerge::hydrate::Value::Map(value) => hydrate_map_to_json(value),
        automerge::hydrate::Value::List(value) => serde_json::Value::Array(
            value
                .iter()
                .map(|value| hydrate_value_to_json(&value.value))
                .collect(),
        ),
        automerge::hydrate::Value::Text(value) => serde_json::json!(value.to_string()),
    }
}

fn scalar_to_json(value: &automerge::ScalarValue) -> serde_json::Value {
    match value {
        automerge::ScalarValue::Bytes(value) => serde_json::json!(value),
        automerge::ScalarValue::Str(value) => serde_json::json!(value),
        automerge::ScalarValue::Int(value) => serde_json::json!(value),
        automerge::ScalarValue::Uint(value) => serde_json::json!(value),
        automerge::ScalarValue::F64(value) => serde_json::json!(value),
        automerge::ScalarValue::Counter(value) => serde_json::json!(i64::from(value)),
        automerge::ScalarValue::Timestamp(value) => serde_json::json!(value),
        automerge::ScalarValue::Boolean(value) => serde_json::json!(value),
        automerge::ScalarValue::Null => serde_json::Value::Null,
        automerge::ScalarValue::Unknown { bytes, .. } => serde_json::json!(bytes),
    }
}

#[no_mangle]
pub extern "C" fn am_text_cursor_position(object_handle: u32, pointer: u32, length: u32) -> i64 {
    let bytes = input_bytes(pointer, length);
    let cursor = match Cursor::try_from(bytes) {
        Ok(cursor) => cursor,
        Err(error) => {
            STATE.with(|state| state.borrow_mut().fail(error));
            return -1;
        }
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let object = match state.object(object_handle) {
            Ok(object) => object,
            Err(error) => {
                state.fail(error);
                return -1;
            }
        };
        match state.doc.get_cursor_position(&object, &cursor, None) {
            Ok(position) => match i64::try_from(position) {
                Ok(position) => {
                    state.error.clear();
                    position
                }
                Err(error) => {
                    state.fail(error);
                    -1
                }
            },
            Err(error) => {
                state.fail(error);
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_commit(
    message_pointer: u32,
    message_length: u32,
    timestamp_seconds: i64,
) -> i32 {
    let message = match input_string(message_pointer, message_length) {
        Ok(message) => message,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let options = CommitOptions::default()
            .with_message(message)
            .with_time(timestamp_seconds);
        match state.doc.commit_with(options) {
            Some(hash) => {
                state.output = hash.as_ref().to_vec();
                state.error.clear();
                0
            }
            None => state.fail("change contains no operations"),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_empty_commit(
    message_pointer: u32,
    message_length: u32,
    timestamp_seconds: i64,
) -> i32 {
    let message = match input_string(message_pointer, message_length) {
        Ok(message) => message,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let options = CommitOptions::default()
            .with_message(message)
            .with_time(timestamp_seconds);
        let hash = state.doc.empty_change(options);
        state.output = hash.as_ref().to_vec();
        state.error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_rollback() -> i64 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        match i64::try_from(state.doc.rollback()) {
            Ok(cancelled) => {
                state.objects.clear();
                state.objects.push(ROOT);
                state.error.clear();
                cancelled
            }
            Err(_) => {
                state.fail("rollback operation count exceeds i64");
                -1
            }
        }
    })
}

#[no_mangle]
pub extern "C" fn am_heads() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        state.output = state
            .doc
            .get_heads()
            .into_iter()
            .flat_map(|hash| hash.0)
            .collect();
        state.error.clear();
    });
    0
}

fn patch_prop_json(prop: &Prop) -> serde_json::Value {
    match prop {
        Prop::Map(key) => serde_json::json!({ "key": key }),
        Prop::Seq(index) => serde_json::json!({ "index": index }),
    }
}

fn patch_value_json(value: &Value<'_>, id: &ObjId) -> Result<serde_json::Value, String> {
    match value {
        Value::Scalar(scalar) => Ok(serde_json::json!({ "scalar": scalar_json(scalar)? })),
        Value::Object(object_type) => Ok(serde_json::json!({
            "object": encode_object_type(*object_type),
            "id": id.to_string(),
        })),
    }
}

fn patches_to_output(patches: &[Patch]) -> Result<Vec<u8>, String> {
    let mut output = Vec::with_capacity(patches.len());
    for patch in patches {
        let object = patch.obj.to_string();
        let action = match &patch.action {
            PatchAction::PutMap {
                key,
                value,
                conflict,
            } => serde_json::json!({
                "type": "put_map",
                "key": key,
                "value": patch_value_json(&value.0, &value.1)?,
                "conflict": conflict,
            }),
            PatchAction::PutSeq {
                index,
                value,
                conflict,
            } => serde_json::json!({
                "type": "put_seq",
                "index": index,
                "value": patch_value_json(&value.0, &value.1)?,
                "conflict": conflict,
            }),
            PatchAction::Insert { index, values } => {
                let mut encoded = Vec::with_capacity(values.len());
                for (value, id, conflict) in values.iter() {
                    encoded.push(serde_json::json!({
                        "value": patch_value_json(value, id)?,
                        "conflict": conflict,
                    }));
                }
                serde_json::json!({ "type": "insert", "index": index, "values": encoded })
            }
            PatchAction::SpliceText {
                index,
                value,
                marks,
            } => {
                let mut action = serde_json::json!({
                    "type": "splice_text",
                    "index": index,
                    "text": value.make_string(),
                });
                if let Some(marks) = marks {
                    let encoded = marks
                        .iter()
                        .map(|(name, value)| {
                            Ok(serde_json::json!({
                                "name": name,
                                "value": scalar_json(value)?,
                            }))
                        })
                        .collect::<Result<Vec<_>, String>>()?;
                    action["marks"] = serde_json::Value::Array(encoded);
                }
                action
            }
            PatchAction::Increment { prop, value } => serde_json::json!({
                "type": "increment",
                "prop": patch_prop_json(prop),
                "value": value,
            }),
            PatchAction::Conflict { prop } => serde_json::json!({
                "type": "conflict",
                "prop": patch_prop_json(prop),
            }),
            PatchAction::DeleteMap { key } => serde_json::json!({
                "type": "delete_map",
                "key": key,
            }),
            PatchAction::DeleteSeq { index, length } => serde_json::json!({
                "type": "delete_seq",
                "index": index,
                "length": length,
            }),
            PatchAction::Mark { marks } => {
                let encoded = marks
                    .iter()
                    .map(|mark| {
                        Ok(serde_json::json!({
                            "start": mark.start,
                            "end": mark.end,
                            "name": mark.name(),
                            "value": scalar_json(mark.value())?,
                        }))
                    })
                    .collect::<Result<Vec<_>, String>>()?;
                serde_json::json!({ "type": "mark", "marks": encoded })
            }
        };
        output.push(serde_json::json!({ "obj": object, "action": action }));
    }

    serde_json::to_vec(&output).map_err(|error| error.to_string())
}

#[no_mangle]
pub extern "C" fn am_diff(
    before_pointer: u32,
    before_length: u32,
    after_pointer: u32,
    after_length: u32,
) -> i32 {
    let before = match input_heads(before_pointer, before_length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    let after = match input_heads(after_pointer, after_length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let patches = state.doc.diff(&before, &after);
        match patches_to_output(&patches) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_update_diff_cursor() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        state.doc.update_diff_cursor();
        state.error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_diff_incremental() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let patches = state.doc.diff_incremental();
        match patches_to_output(&patches) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_current_state() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let patches = state.doc.document().current_state();
        match patches_to_output(&patches) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_stats() -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let stats = state.doc.stats();
        let value = serde_json::json!({
            "numChanges": stats.num_changes,
            "numOps": stats.num_ops,
            "numActors": stats.num_actors,
        });
        match serde_json::to_vec(&value) {
            Ok(output) => {
                state.output = output;
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_has_heads(pointer: u32, length: u32) -> i32 {
    let heads = match input_heads(pointer, length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let has_heads = heads
            .iter()
            .all(|head| state.doc.get_change_by_hash(head).is_some());
        state.error.clear();
        if has_heads {
            1
        } else {
            0
        }
    })
}

#[no_mangle]
pub extern "C" fn am_missing_dependencies(pointer: u32, length: u32) -> i32 {
    let heads = match input_heads(pointer, length) {
        Ok(heads) => heads,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        state.output = state
            .doc
            .get_missing_deps(&heads)
            .into_iter()
            .flat_map(|hash| hash.0)
            .collect();
        state.error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_merge(pointer: u32, length: u32) -> i32 {
    let bytes = input_bytes(pointer, length);
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let mut other = match AutoCommit::load(&bytes) {
            Ok(other) => other,
            Err(error) => return state.fail(error),
        };
        match state.doc.merge(&mut other) {
            Ok(heads) => {
                state.output = heads.into_iter().flat_map(|hash| hash.0).collect();
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_sync_new() -> i64 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let handle = match u32::try_from(state.sync_states.len()) {
            Ok(handle) => handle,
            Err(_) => {
                state.fail("too many sync states");
                return -1;
            }
        };
        state.sync_states.push(Some(SyncState::new()));
        state.error.clear();
        i64::from(handle)
    })
}

#[no_mangle]
pub extern "C" fn am_sync_set_read_only(handle: u32, read_only: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let Some(sync_state) = state
            .sync_states
            .get_mut(handle as usize)
            .and_then(Option::as_mut)
        else {
            return state.fail(format!("invalid sync state {handle}"));
        };
        sync_state.set_read_only(read_only != 0);
        state.error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_sync_peer_read_only(handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let Some(sync_state) = state
            .sync_states
            .get(handle as usize)
            .and_then(Option::as_ref)
        else {
            return state.fail(format!("invalid sync state {handle}"));
        };
        if sync_state.is_peer_read_only() {
            1
        } else {
            0
        }
    })
}

#[no_mangle]
pub extern "C" fn am_sync_free(handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let Some(sync_state) = state.sync_states.get_mut(handle as usize) else {
            return state.fail(format!("invalid sync state handle {handle}"));
        };
        if sync_state.take().is_none() {
            return state.fail(format!("sync state handle {handle} is closed"));
        }
        state.error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_sync_generate(handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let State {
            doc,
            sync_states,
            output,
            error,
            ..
        } = &mut *state;
        let Some(Some(sync_state)) = sync_states.get_mut(handle as usize) else {
            return state.fail(format!("invalid sync state handle {handle}"));
        };

        *output = doc
            .sync()
            .generate_sync_message(sync_state)
            .map(Message::encode)
            .unwrap_or_default();
        error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_sync_receive(handle: u32, pointer: u32, length: u32) -> i32 {
    let bytes = input_bytes(pointer, length);
    let message = match Message::decode(&bytes) {
        Ok(message) => message,
        Err(error) => return STATE.with(|state| state.borrow_mut().fail(error)),
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        if !matches!(state.sync_states.get(handle as usize), Some(Some(_))) {
            return state.fail(format!("invalid sync state handle {handle}"));
        }

        let result = {
            let State {
                doc, sync_states, ..
            } = &mut *state;
            let sync_state = sync_states[handle as usize].as_mut().unwrap();
            doc.sync().receive_sync_message(sync_state, message)
        };
        match result {
            Ok(()) => {
                state.error.clear();
                0
            }
            Err(error) => state.fail(error),
        }
    })
}

#[no_mangle]
pub extern "C" fn am_sync_save(handle: u32) -> i32 {
    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let Some(Some(sync_state)) = state.sync_states.get(handle as usize) else {
            return state.fail(format!("invalid sync state handle {handle}"));
        };
        state.output = sync_state.encode();
        state.error.clear();
        0
    })
}

#[no_mangle]
pub extern "C" fn am_sync_load(pointer: u32, length: u32) -> i64 {
    let bytes = input_bytes(pointer, length);
    let sync_state = match SyncState::decode(&bytes) {
        Ok(sync_state) => sync_state,
        Err(error) => {
            STATE.with(|state| state.borrow_mut().fail(error));
            return -1;
        }
    };

    STATE.with(|state| {
        let mut state = state.borrow_mut();
        let handle = match u32::try_from(state.sync_states.len()) {
            Ok(handle) => handle,
            Err(_) => {
                state.fail("too many sync states");
                return -1;
            }
        };
        state.sync_states.push(Some(sync_state));
        state.error.clear();
        i64::from(handle)
    })
}
