
package logging

func GetLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	zapLogger, err := config.Build()
}

