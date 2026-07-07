package core

import "testing"

func TestCheckImageSourceAccessSkipsRegistryProbeForLocalImage(t *testing.T) {
	calls := 0
	sources := checkImageSourceAccess([]ImageSource{{
		Service:    "web",
		Image:      "nginx:latest",
		SourceType: "registry",
		Registry:   "docker.io",
	}}, imageSourceProbeFuncs{
		imagePresentLocally: func(image string) bool {
			if image != "nginx:latest" {
				t.Fatalf("unexpected local image check for %q", image)
			}
			return true
		},
		hasStoredAuthForRegistry: func(string) bool {
			calls++
			return false
		},
		manifestInspect: func(string, bool) (bool, string) {
			calls++
			return false, "should not be called"
		},
	})

	if calls != 0 {
		t.Fatalf("expected no auth or manifest probes for local image, got %d", calls)
	}
	if got := sources[0].Access; got != "local" {
		t.Fatalf("expected local access, got %q", got)
	}
}

func TestCheckImageSourceAccessUsesOnlyAuthenticatedProbeWhenLoggedIn(t *testing.T) {
	var authProbes, anonymousProbes int
	sources := checkImageSourceAccess([]ImageSource{{
		Service:    "api",
		Image:      "private.example.com/team/api:1.0",
		SourceType: "registry",
		Registry:   "private.example.com",
	}}, imageSourceProbeFuncs{
		imagePresentLocally: func(string) bool { return false },
		hasStoredAuthForRegistry: func(registry string) bool {
			return registry == "private.example.com"
		},
		manifestInspect: func(image string, anonymous bool) (bool, string) {
			if anonymous {
				anonymousProbes++
			} else {
				authProbes++
			}
			if image != "private.example.com/team/api:1.0" {
				t.Fatalf("unexpected manifest inspect for %q", image)
			}
			return true, ""
		},
	})

	if authProbes != 1 {
		t.Fatalf("expected one authenticated probe, got %d", authProbes)
	}
	if anonymousProbes != 0 {
		t.Fatalf("expected no anonymous probe when auth exists, got %d", anonymousProbes)
	}
	if got := sources[0].Access; got != "authenticated" {
		t.Fatalf("expected authenticated access, got %q", got)
	}
}

func TestCheckImageSourceAccessUsesAnonymousProbeWithoutStoredAuth(t *testing.T) {
	var authProbes, anonymousProbes int
	sources := checkImageSourceAccess([]ImageSource{{
		Service:    "web",
		Image:      "nginx:latest",
		SourceType: "registry",
		Registry:   "docker.io",
	}}, imageSourceProbeFuncs{
		imagePresentLocally:      func(string) bool { return false },
		hasStoredAuthForRegistry: func(string) bool { return false },
		manifestInspect: func(_ string, anonymous bool) (bool, string) {
			if anonymous {
				anonymousProbes++
			} else {
				authProbes++
			}
			return true, ""
		},
	})

	if anonymousProbes != 1 {
		t.Fatalf("expected one anonymous probe, got %d", anonymousProbes)
	}
	if authProbes != 0 {
		t.Fatalf("expected no authenticated probe without stored auth, got %d", authProbes)
	}
	if got := sources[0].Access; got != "public" {
		t.Fatalf("expected public access, got %q", got)
	}
}

func TestCheckImageSourceAccessDoesNotAnonymousProbeAfterAuthFailure(t *testing.T) {
	var anonymousProbes int
	sources := checkImageSourceAccess([]ImageSource{{
		Service:    "api",
		Image:      "private.example.com/team/api:missing",
		SourceType: "registry",
		Registry:   "private.example.com",
	}}, imageSourceProbeFuncs{
		imagePresentLocally:      func(string) bool { return false },
		hasStoredAuthForRegistry: func(string) bool { return true },
		manifestInspect: func(_ string, anonymous bool) (bool, string) {
			if anonymous {
				anonymousProbes++
			}
			return false, "unauthorized: authentication required"
		},
	})

	if anonymousProbes != 0 {
		t.Fatalf("expected auth failure to skip anonymous probe, got %d anonymous probes", anonymousProbes)
	}
	if got := sources[0].Access; got != "private-login-required" {
		t.Fatalf("expected private-login-required access, got %q", got)
	}
}
