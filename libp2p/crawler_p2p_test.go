package libp2p

//func TestCrawler_connect(t *testing.T) {
//	ctx := context.Background()
//	clck := clock.NewMock()
//	ctrl := gomock.NewController(t)
//
//	pid := peer.AddrInfo{
//		ID: "test-peer",
//		Addrs: []ma.Multiaddr{
//			nebtest.MustMultiaddr(t, "/ip4/123.123.123.123/tcp/1234"),
//		},
//	}
//	crawlCfg := DefaultCrawlerConfig()
//	crawlCfg.Clock = clck
//
//	tests := []struct {
//		name        string
//		connectErrs []error
//		expectedErr error
//	}{
//		{
//			name:        "success",
//			connectErrs: []error{nil},
//		},
//		{
//			name:        "unhandled",
//			connectErrs: []error{fmt.Errorf("unhandled")},
//			expectedErr: fmt.Errorf("unhandled"),
//		},
//		{
//			name: "connection refused",
//			connectErrs: []error{
//				fmt.Errorf("connection refused"),
//				fmt.Errorf("connection refused"),
//				fmt.Errorf("connection refused"),
//			},
//			expectedErr: fmt.Errorf("connection refused"),
//		},
//		{
//			name: "connection refused then good",
//			connectErrs: []error{
//				fmt.Errorf("connection refused"),
//				nil,
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			h := NewMockHost(ctrl)
//			crawler := &Crawler{
//				id:   "test-01",
//				cfg:  crawlCfg,
//				host: h,
//			}
//
//			callCount := 0
//			h.EXPECT().
//				Connect(gomock.Any(), pid).
//				Times(len(tt.connectErrs)).
//				DoAndReturn(func(context.Context, peer.AddrInfo) error {
//					err := tt.connectErrs[callCount]
//					callCount += 1
//					return err
//				})
//
//			err := crawler.connect(ctx, pid)
//			if tt.expectedErr == nil {
//				assert.NoError(t, err)
//			} else {
//				assert.ErrorContains(t, err, tt.expectedErr.Error())
//			}
//		})
//	}
//}
