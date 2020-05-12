// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package bindingenvironment provides an ltl.Environment and helpers to
// support binding names to values.  See docs/binding.md.
package bindingenvironment

import (
    // "fmt"
    "ltl/pkg/bindings"
    "ltl/pkg/ltl"
    "ltl/pkg/tags"
)

// bindingEnvironment describes an Environment capable of binding values to
// names.
type bindingEnvironment interface {
    ltl.Environment
    // Bindings returns the set of Bindings in this Environment.  Bindings are
    // only provided by matching Environments.
    bindings() *bindings.Bindings
    // Tags returns the set of Tags associated with this Environment.  Tags are
    // only provided by matching Environments.
    tags() *tags.Tags
    // hasReference returns true iff this bindingEnvironment contains
    // references, either directly or indirectly.
    hasReferences() bool
    // applyBindings returns a new ltl.Environment resulting from binding the
    // provided Bindings in the receiver.  applyBindings should simplify the
    // tree wherever possible, e.g. by demoting an intermediateNode to a
    // State.
    applyBindings(bindings *bindings.Bindings) ltl.Environment
    // merge attempts to merge the receiver with the argument, in order to
    // reduce the size of the bindingEnvironment tree and process Matches more
    // efficiently.  Two Environments may merge if they are both
    // bindingEnvironments of the same type, their matching state is equivalent
    // (matching, bindings, and references all equal), and their children are
    // equivalent (possibly with a different order.)  If the two Environments
    // are not equivalent, merge returns false, meaning the Environments cannot
    // be merged.
    // Tags are not considered for equivalence, and the returned Environment's
    // Tags are the union of the two arguments'.
    merge(oe ltl.Environment) (bindingEnvironment, bool)
}

// Bindings returns the set of Bindings bound by the provided Environment.  If
// the provided Environment is not binding, a nil Bindings is returned.
func Bindings(env ltl.Environment) *bindings.Bindings {
    if be, ok := env.(bindingEnvironment); ok {
        return be.bindings()
    }
    return nil
}

// Tags returns the set of Tags tagged by the provided Environment.  If the
// provided Environment is not binding, a nil Tags is returned.
func Tags(env ltl.Environment) *tags.Tags {
    if be, ok := env.(bindingEnvironment); ok {
        return be.tags()
    }
    return nil
}

// Helper functions to safely handle Environments that may not be binding.

func hasReferences(env ltl.Environment) bool {
    if be, ok := env.(bindingEnvironment); ok {
        return be.hasReferences()
    }
    return false
}

func applyBindings(b *bindings.Bindings, env ltl.Environment) ltl.Environment {
    if be, ok := env.(bindingEnvironment); ok {
        return be.applyBindings(b)
    }
    return env
}

func matchIfSatisfied(env ltl.Environment) bool {
    if be, ok := env.(*bindingNode); ok {
        return be.matching
    }
    return env.Matching()
}

func merge(a, b ltl.Environment) (bindingEnvironment, bool) {
    if be, ok := a.(bindingEnvironment); ok {
        return be.merge(b)
    }
    return nil, false
}
