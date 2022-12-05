/*
Copyright 2022 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/kcloutie/cluster-interceptors/add-changed-files/pkg"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"github.com/tektoncd/triggers/pkg/interceptors/server"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
)

const (
	// Port is the port that the port that interceptor service listens on
	Port         = 8082
	readTimeout  = 5 * time.Second
	writeTimeout = 20 * time.Second
	idleTimeout  = 60 * time.Second

	// authSecretEnvVar = "GITHUB_OAUTH_SECRET"
)

func main() {
	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.NewContext()
	cfg := injection.ParseAndGetRESTConfigOrDie()

	// ctx, startInformer := injection.EnableInjectionOrDie(ctx, cfg)
	ctx, _ = injection.EnableInjectionOrDie(ctx, cfg)

	logger := logging.FromContext(ctx)

	s := server.Server{
		Logger: logger,
	}
	s.RegisterInterceptor("addChangedFiles", pkg.Interceptor{
		SecretGetter: interceptors.DefaultSecretGetter(kubeclient.Get(ctx).CoreV1()),
	})

	mux := http.NewServeMux()
	mux.Handle("/", &s)
	mux.HandleFunc("/ready", handler)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", Port),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      mux,
	}

	logger.Infof("Listen and serve on port %d", Port)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("failed to start interceptors service: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
