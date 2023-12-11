If you want to validate your aggregatedSelection functions:

1. create a sub-directory
2. create a file named __cleaner.yaml__ containing your Cleaner instance
3. create a file named __resources.yaml__ containing all the resource Cleaner instance will find
4. create a file named __matching.yaml__ containing all the resources that matches __AggregatedSelection__
5. run ``make test``

That will run the exact code Cleaner will run in your cluster. 
If you see no error, your Cleaner instance is correct