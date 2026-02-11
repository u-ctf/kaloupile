package kind

func DeleteCluster() error {
	return deleteCluster(ClusterName)
}

func deleteCluster(name string) error {
	_, err := runKindLogged("delete", "cluster", "--name", name)
	return err
}
