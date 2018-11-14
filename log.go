package jrpc2

// InfoLogf convinience wrapper for info logger
func (s *Service) InfoLogf(format string, v ...interface{}) {
	s.InfoLogger.Printf(format, v...)
}

// InfoLogln convinience wrapper for info logger
func (s *Service) InfoLogln(v ...interface{}) {
	s.InfoLogger.Println(v...)
}

// InfoLog convinience wrapper for info logger
func (s *Service) InfoLog(v ...interface{}) {
	s.InfoLogger.Print(v...)
}

// ErrorLogf convinience wrapper for error logger
func (s *Service) ErrorLogf(format string, v ...interface{}) {
	s.ErrorLogger.Printf(format, v...)
}

// ErrorLogln convinience wrapper for error logger
func (s *Service) ErrorLogln(v ...interface{}) {
	s.ErrorLogger.Println(v...)
}

// ErrorLog convinience wrapper for error logger
func (s *Service) ErrorLog(v ...interface{}) {
	s.ErrorLogger.Print(v...)
}

// CriticalLogf convinience wrapper for critical logger
func (s *Service) CriticalLogf(format string, v ...interface{}) {
	s.CriticalLogger.Printf(format, v...)
}

// CriticalLogln convinience wrapper for critical logger
func (s *Service) CriticalLogln(v ...interface{}) {
	s.CriticalLogger.Println(v...)
}

// CriticalLog convinience wrapper for critical logger
func (s *Service) CriticalLog(v ...interface{}) {
	s.CriticalLogger.Print(v...)
}

// Panicf convinience wrapper for critical panic logger
func (s *Service) Panicf(format string, v ...interface{}) {
	s.CriticalLogger.Panicf(format, v...)
}

// Panicln convinience wrapper for critical panic logger
func (s *Service) Panicln(v ...interface{}) {
	s.CriticalLogger.Panicln(v...)
}

// Panic convinience wrapper for critical panic logger
func (s *Service) Panic(v ...interface{}) {
	s.CriticalLogger.Panic(v...)
}

// Fatalf convinience wrapper for critical fatal logger
func (s *Service) Fatalf(format string, v ...interface{}) {
	s.CriticalLogger.Fatalf(format, v...)
}

// Fatalln convinience wrapper for critical fatal logger
func (s *Service) Fatalln(v ...interface{}) {
	s.CriticalLogger.Fatalln(v...)
}

// Fatal convinience wrapper for critical fatal logger
func (s *Service) Fatal(v ...interface{}) {
	s.CriticalLogger.Fatal(v...)
}
