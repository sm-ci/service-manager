package filters

import (
	"context"
	"fmt"
	"github.com/Peripli/service-manager/pkg/log"
	"github.com/Peripli/service-manager/pkg/types"
	"github.com/Peripli/service-manager/pkg/util"
	"github.com/Peripli/service-manager/pkg/util/slice"
	"github.com/Peripli/service-manager/pkg/web"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"net/http"
	"strings"
)

type RateLimiterFilter struct {
	rateLimiters        []*stdlib.Middleware
	excludeList         []string
	excludeByPrefixList []string
	tenantLabelKey      string
}

func NewRateLimiterFilter(middleware []*stdlib.Middleware, excludeList, excludeByPrefixList []string, tenantLabelKey string) *RateLimiterFilter {
	return &RateLimiterFilter{
		rateLimiters:        middleware,
		excludeList:         excludeList,
		tenantLabelKey:      tenantLabelKey,
		excludeByPrefixList: excludeByPrefixList,
	}
}

func (rl *RateLimiterFilter) handleLimitIsReached(limiterContext limiter.Context, username string, isLimitedClient bool, context context.Context) error {
	if !isLimitedClient {
		return nil
	}

	log.C(context).Debugf("Request limit has been exceeded for client with key", username)
	return &util.HTTPError{
		ErrorType:   "BadRequest",
		Description: fmt.Sprintf("The allowed request limit of %s requests has been reached please try again later", limiterContext.Limit),
		StatusCode:  http.StatusTooManyRequests,
	}
}

func (rl *RateLimiterFilter) isRateLimitedClient(userContext *web.UserContext) (bool, error) {
	//don't restrict global users
	if userContext.AccessLevel == web.GlobalAccess || userContext.AccessLevel == web.AllTenantAccess {
		return false, nil
	}

	excludeByName := userContext.Name
	if userContext.AuthenticationType == web.Basic {
		platform := types.Platform{}
		err := userContext.Data(&platform)
		if err != nil {
			return false, err
		}

		if _, isTenantScopedPlatform := platform.Labels[rl.tenantLabelKey]; !isTenantScopedPlatform {
			return false, nil
		}

		excludeByName = platform.Name
	}

	if slice.StringsAnyEquals(rl.excludeList, excludeByName) {
		return false, nil
	}

	return true, nil
}

func (rl *RateLimiterFilter) Name() string {
	return "RateLimiterFilter"
}

func (rl *RateLimiterFilter) isAllowedPrefix(path string) bool {
	for _, excludedPathPrefix := range rl.excludeByPrefixList {
		if strings.HasPrefix(excludedPathPrefix, path) {
			return false
		}
	}
	return true
}

func (rl *RateLimiterFilter) Run(request *web.Request, next web.Handler) (*web.Response, error) {
	userContext, isProtectedEndpoint := web.UserFromContext(request.Context())

	if !isProtectedEndpoint || rl.isAllowedPrefix(request.URL.Path) {
		//skip public endpoints or excluded prefix's
		return next.Handle(request)
	}

	isLimitedClient, err := rl.isRateLimitedClient(userContext)
	if err != nil {
		log.C(request.Context()).WithError(err).Errorf("unable to determine if client should be rate limited")
		return nil, err
	}

	for _, rlm := range rl.rateLimiters {
		limiterContext, err := rlm.Limiter.Get(request.Context(), userContext.Name)
		if err != nil {
			return nil, err
		}

		// Log the clients that reach half of the allowed limit
		if limiterContext.Remaining == limiterContext.Limit/10 {
			log.C(request.Context()).Infof("the client has already used 10% of it's allowed requests, is_limited_client:%s,client key:%s, X-RateLimit-Limit=%s,X-o-Remaining=%s,X-RateLimit-Reset=%s", isLimitedClient, userContext.Name, limiterContext.Limit, limiterContext.Reset)
		}

		if limiterContext.Reached {
			err := rl.handleLimitIsReached(limiterContext, userContext.Name, isLimitedClient, request.Context())
			if err != nil {
				return nil, err
			}
		}
	}

	return next.Handle(request)
}

func (rl *RateLimiterFilter) FilterMatchers() []web.FilterMatcher {
	return []web.FilterMatcher{
		{
			Matchers: []web.Matcher{
				web.Path("/**"),
				web.Methods(http.MethodPost, http.MethodPatch, http.MethodGet, http.MethodDelete, http.MethodOptions),
			},
		},
	}
}