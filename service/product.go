package service

import (
	"context"
	logging "github.com/sirupsen/logrus"
	"mall/dao"
	"mall/model"
	"mall/pkg/e"
	"mall/serializer"
	"mime/multipart"
	"strconv"
	"sync"
)

//更新商品的服务
type ProductService struct {
	ID            uint   `form:"id" json:"id"`
	Name          string `form:"name" json:"name"`
	CategoryID    int    `form:"category_id" json:"category_id"`
	Title         string `form:"title" json:"title" binding:"required,min=2,max=100"`
	Info          string `form:"info" json:"info" binding:"max=1000"`
	ImgPath       string `form:"img_path" json:"img_path"`
	Price         string `form:"price" json:"price"`
	DiscountPrice string `form:"discount_price" json:"discount_price"`
	OnSale        bool   `form:"on_sale" json:"on_sale"`
	Num           int    `form:"num" json:"num"`
	model.BasePage
}

type ListProductImgService struct {
}

// 商品
func (service *ProductService) Show(ctx context.Context, id string) serializer.Response {
	code := e.SUCCESS

	pId, _ := strconv.Atoi(id)

	productDao := dao.NewProductDao(ctx)
	product, err := productDao.GetProductById(uint(pId))
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Status: code,
		Data:   serializer.BuildProduct(product),
		Msg:    e.GetMsg(code),
	}
}

//创建商品
func (service *ProductService) Create(ctx context.Context, uId uint, files []*multipart.FileHeader) serializer.Response {
	var boss model.User
	var err error
	code := e.SUCCESS

	userDao := dao.NewUserDao(ctx)
	boss, _ = userDao.GetUserById(uId)
	// 以第一张作为封面图
	tmp, _ := files[0].Open()
	status, info := UploadToQiNiu(tmp, files[0].Size)
	if status != 200 {
		return serializer.Response{
			Status: status,
			Data:   e.GetMsg(status),
			Error:  info,
		}
	}
	product := model.Product{
		Name:          service.Name,
		CategoryID:    uint(service.CategoryID),
		Title:         service.Title,
		Info:          service.Info,
		ImgPath:       info,
		Price:         service.Price,
		DiscountPrice: service.DiscountPrice,
		Num:           service.Num,
		OnSale:        true,
		BossID:        int(uId),
		BossName:      boss.UserName,
		BossAvatar:    boss.Avatar,
	}
	productDao := dao.NewProductDao(ctx)
	err = productDao.CreateProduct(product)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(files))
	for _, file := range files {
		productImgDao := dao.NewProductImgDaoByDB(productDao.DB)
		tmp, _ := file.Open()
		status, info := UploadToQiNiu(tmp, file.Size)
		if status != 200 {
			return serializer.Response{
				Status: status,
				Data:   e.GetMsg(status),
				Error:  info,
			}
		}
		productImg := model.ProductImg{
			ProductID: product.ID,
			ImgPath:   info,
		}
		err = productImgDao.CreateProductImg(productImg)
		if err != nil {
			code = e.ERROR
			return serializer.Response{
				Status: code,
				Msg:    e.GetMsg(code),
				Error:  err.Error(),
			}
		}
		wg.Done()
	}

	wg.Wait()

	return serializer.Response{
		Status: code,
		Data:   serializer.BuildProduct(product),
		Msg:    e.GetMsg(code),
	}
}

func (service *ProductService) List(ctx context.Context) serializer.Response {
	var products []model.Product
	var total int64
	code := e.SUCCESS

	if service.PageSize == 0 {
		service.PageSize = 15
	}
	condition := make(map[string]interface{})
	if service.CategoryID != 0 {
		condition["category_id"] = service.CategoryID
	}
	productDao := dao.NewProductDao(ctx)
	total, err := productDao.CountProductByCondition(condition)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		productDao = dao.NewProductDaoByDB(productDao.DB)
		products, _ = productDao.ListProductByCondition(condition, service.BasePage)
		wg.Done()
	}()
	wg.Wait()

	var productList []serializer.Product
	for _, item := range products {
		products := serializer.BuildProduct(item)
		productList = append(productList, products)
	}

	return serializer.BuildListResponse(serializer.BuildProducts(products), uint(total))
}

//删除商品
func (service *ProductService) Delete(ctx context.Context, pId string) serializer.Response {
	code := e.SUCCESS

	productDao := dao.NewProductDao(ctx)
	productId, _ := strconv.Atoi(pId)
	err := productDao.DeleteProduct(uint(productId))
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// 更新商品
func (service *ProductService) Update(ctx context.Context, pId string) serializer.Response {
	code := e.SUCCESS
	productDao := dao.NewProductDao(ctx)

	productId, _ := strconv.Atoi(pId)
	product := model.Product{
		Name:          service.Name,
		CategoryID:    uint(service.CategoryID),
		Title:         service.Title,
		Info:          service.Info,
		ImgPath:       service.ImgPath,
		Price:         service.Price,
		DiscountPrice: service.DiscountPrice,
		OnSale:        service.OnSale,
	}
	err := productDao.UpdateProduct(uint(productId), product)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// 搜索商品
func (service *ProductService) Search(ctx context.Context) serializer.Response {
	code := e.SUCCESS
	if service.PageSize == 0 {
		service.PageSize = 15
	}

	productDao := dao.NewProductDao(ctx)
	products, err := productDao.SearchProduct(service.Info, service.BasePage)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.BuildListResponse(serializer.BuildProducts(products), uint(len(products)))
}

// List 获取商品列表图片
func (service *ListProductImgService) List(ctx context.Context, pId string) serializer.Response {
	productImgDao := dao.NewProductImgDao(ctx)
	productId, _ := strconv.Atoi(pId)
	productImgs, _ := productImgDao.ListProductImgByProductId(uint(productId))
	return serializer.BuildListResponse(serializer.BuildProductImgs(productImgs), uint(len(productImgs)))
}
