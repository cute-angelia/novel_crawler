package internal

func retry(task func() error, count int) error {
	if err := task(); err == nil {
		return nil
	} else if count > 1 {
		//log.Println("do retry, remain retry count:", count-1)
		return retry(task, count-1)
	} else {
		return err
	}

}
