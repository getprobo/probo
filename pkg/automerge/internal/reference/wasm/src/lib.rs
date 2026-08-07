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

use automerge::sync::{Message, State as SyncState, SyncDoc};
use automerge::transaction::{CommitOptions, Transactable};
use automerge::{ActorId, AutoCommit, Cursor, ObjId, ObjType, ReadDoc, Value, ROOT};

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
