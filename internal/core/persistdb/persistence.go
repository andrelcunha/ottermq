package persistdb

// func (b *VHost) saveMessage(queueName string, msg Message) error {
// wd, err := os.Getwd()
// dir := filepath.Join(wd, "data", "queues", queueName)
// if err := os.MkdirAll(dir, 0755); err != nil {
// 	return err
// }
// file := filepath.Join(dir, msg.ID+".json")
// data, err := json.Marshal(msg)
// if err != nil {
// 	return err
// }
// return os.WriteFile(file, data, 0644)
// 	return nil
// }

// func (b *VHost) loadMessages(queueName string) ([]Message, error) {
// dir := filepath.Join("data", "queues", queueName)
// files, err := os.ReadDir(dir)
// if err != nil {
// 	return nil, err
// }
// var messages []Message
// for _, file := range files {
// 	data, err := os.ReadFile(filepath.Join(dir, file.Name()))
// 	if err != nil {
// 		return nil, err
// 	}
// 	var msg Message
// 	if err := json.Unmarshal(data, &msg); err != nil {
// 		return nil, err
// 	}
// 	messages = append(messages, msg)
// }
// return messages, nil
// return nil, nil
// }

// func (b *VHost) saveBrokerState() error {
// data, err := json.Marshal(b)
// if err != nil {
// 	return err
// }
// // get the current directory
// wd, err := os.Getwd()
// brokerFile := filepath.Join(wd, "data", "broker_state.json")
// err = os.MkdirAll(filepath.Dir(brokerFile), 0755)
// return os.WriteFile(brokerFile, data, 0644)
// return nil
// }

// func (b *VHost) loadBrokerState() error {
// brokerFile := filepath.Join("data", "broker_state.json")
// data, err := os.ReadFile(brokerFile)
// if err != nil {
// 	return err
// }
// return json.Unmarshal(data, b)
// return nil
// }
